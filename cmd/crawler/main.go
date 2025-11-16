package main

import (
	"context"
	"fmt"
	"temperature-checker/internal/config"
	"temperature-checker/internal/core/crawler"
	"temperature-checker/internal/core/meteo"
	"temperature-checker/internal/db"
	"temperature-checker/internal/logger"
	"temperature-checker/internal/mqtt"
)

func main() {
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
		panic(fmt.Errorf("failed to create database connection: %w", err))
	}

	meteoClient := meteo.NewOpenMeteoClient(&meteo.OpenMeteoDependencies{})

	broker, err := mqtt.NewMosquittoClient(mqtt.Dependencies{
		Logger: log,
		Config: &cfg.MQTTBroker,
	})

	if err != nil {
		panic(fmt.Errorf("failed to create mqtt client: %w", err))
	}

	defer broker.Close()

	crawlerService := crawler.NewService(&crawler.ServiceDependencies{
		DB:          conManager,
		Logger:      log,
		MeteoClient: meteoClient,
		Broker:      broker,
	})

	rootCtx := context.Background()
	defer rootCtx.Done()

	if err := crawlerService.Crawl(rootCtx); err != nil {
		log.Error("failed to crawl", err)
	}
}
