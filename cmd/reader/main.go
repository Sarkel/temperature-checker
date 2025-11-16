package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"temperature-checker/internal/config"
	"temperature-checker/internal/core/reader"
	"temperature-checker/internal/db"
	"temperature-checker/internal/logger"
	"temperature-checker/internal/mqtt"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()

	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	log := logger.New(logger.Dependencies{
		Config: cfg.Logger,
	})

	conManager, err := db.NewConManager(db.Dependencies{
		Logger: log,
		Config: &cfg.Database,
	})

	defer db.Close(conManager, log)

	if err != nil {
		log.Error("failed to create database connection", "err", err)
	}

	broker, err := mqtt.NewMosquittoClient(mqtt.Dependencies{
		Logger: log,
		Config: &cfg.MQTTBroker,
	})

	if err != nil {
		log.Error("failed to create mqtt client", "err", err)
	}

	defer broker.Close()

	readerService := reader.NewService(&reader.Dependencies{
		DB:     conManager,
		Logger: log,
		Broker: broker,
	})

	if err := readerService.Listen(ctx); err != nil {
		log.Error("failed to listen to mqtt broker", "err", err)
	}

	log.Info("reader service running...")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh // wait for Ctrl+C / SIGTERM

	log.Info("reader service stopped")
}
