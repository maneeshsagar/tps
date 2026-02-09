package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/maneeshsagar/tps/config"
	"github.com/maneeshsagar/tps/internal/adapters/repository"
	"github.com/maneeshsagar/tps/internal/application"
	"github.com/maneeshsagar/tps/internal/core/ports"
	"github.com/maneeshsagar/tps/internal/infrastructure"
	"github.com/maneeshsagar/tps/logger"
)

func main() {

	// initialize logger first to capture any logs during startup
	log := logger.Default()

	config.LoadEnv(log)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config", "err", err)
	}

	// reinitialize logger with config level
	log = logger.NewZeroLogger(cfg.Log.Level)

	db, err := infrastructure.NewGormPostgres(cfg.Postgres, log)
	if err != nil {
		log.Fatal(err.Error(), "err", err)
	}

	// repositories
	accountRepo := repository.NewAccountRepo(db)
	txnRepo := repository.NewTransactionRepo(db)
	asyncTxRepo := repository.NewAsyncTransactionRepo(db)

	// infrastructure
	txManager := repository.NewTxManager(db)
	lockManager := infrastructure.NewLockManager(db)
	kafkaProducer := infrastructure.NewKafkaProducer(cfg.Kafka.Brokers)
	kafkaConsumer := infrastructure.NewKafkaConsumer(cfg.Kafka.Brokers, log)
	defer kafkaProducer.Close()

	// service
	svc := application.NewTransferService(
		accountRepo, txnRepo, asyncTxRepo,
		txManager, lockManager, kafkaProducer, log,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info("shutting down...")
		cancel()
	}()

	log.Info("consumer starting...")
	err = kafkaConsumer.Subscribe(ctx, application.TopicTransactions, func(msg ports.Message) error {
		var tm application.TransferMessage
		if err := json.Unmarshal(msg.Value, &tm); err != nil {
			log.Error("failed to unmarshal message", "err", err)
			return nil
		}
		return svc.ProcessTransfer(ctx, tm)
	})
	if err != nil && err != context.Canceled {
		log.Fatal("consumer failed", "err", err)
	}
	log.Info("consumer stopped")
}
