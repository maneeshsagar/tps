package main

import (
	"fmt"

	"github.com/maneeshsagar/tps/config"
	"github.com/maneeshsagar/tps/internal/adapters/http"
	"github.com/maneeshsagar/tps/internal/adapters/repository"
	"github.com/maneeshsagar/tps/internal/application"
	"github.com/maneeshsagar/tps/internal/infrastructure"
	"github.com/maneeshsagar/tps/logger"
)

func main() {
	log := logger.Default()

	config.LoadEnv(log)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config", "err", err)
	}

	log = logger.NewZeroLogger(cfg.Log.Level)

	db, err := infrastructure.NewGormPostgres(cfg.Postgres, log)
	if err != nil {
		log.Fatal("failed to connect postgres", "err", err)
	}

	// only app runs migrations
	if err := infrastructure.RunMigrations(db, log); err != nil {
		log.Fatal("migration failed", "err", err)
	}

	// repositories
	accountRepo := repository.NewAccountRepo(db)
	txnRepo := repository.NewTransactionRepo(db)
	asyncTxRepo := repository.NewAsyncTransactionRepo(db)

	// infrastructure
	txManager := repository.NewTxManager(db)
	lockManager := infrastructure.NewLockManager(db)
	kafkaProducer := infrastructure.NewKafkaProducer(cfg.Kafka.Brokers)
	defer kafkaProducer.Close()

	// service (includes sync + async transfer)
	svc := application.NewTransferService(
		accountRepo, txnRepo, asyncTxRepo,
		txManager, lockManager, kafkaProducer, log,
	)

	router := http.NewRouter(svc)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Info("server starting", "addr", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("server failed", "err", err)
	}
}
