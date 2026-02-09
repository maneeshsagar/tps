package infrastructure

import (
	"fmt"
	"time"

	"github.com/maneeshsagar/tps/config"
	"github.com/maneeshsagar/tps/internal/adapters/repository"
	"github.com/maneeshsagar/tps/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewGormPostgres(cfg config.PostgresConfig, log logger.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime())

	done := make(chan error, 1)
	go func() { done <- sqlDB.Ping() }()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("postgres ping failed: %w", err)
		}
	case <-time.After(2 * time.Second):
		return nil, fmt.Errorf("postgres ping timeout")
	}

	log.Info("connected to postgres")
	return db, nil
}

// RunMigrations should only be called by the app server, not the consumer
func RunMigrations(db *gorm.DB, log logger.Logger) error {
	err := db.AutoMigrate(
		&repository.AccountModel{},
		&repository.TransactionModel{},
		&repository.AsyncTransactionStatusModel{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Info("database schema migrated")
	return nil
}
