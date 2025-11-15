package main

import (
	"context"
	"fmt"
	"log/slog"
	"temperature-checker/internal/config"
	"temperature-checker/internal/core/crawler"
	"temperature-checker/internal/core/meteo"
	"temperature-checker/internal/db"
	"temperature-checker/internal/logger"
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

	defer func(conManager *db.ConManager, log *slog.Logger) {
		if err := conManager.Close(); err != nil {
			log.Error("failed to close database connection", err)
		}
	}(conManager, log)

	if err != nil {
		log.Error("failed to create database connection", err)
	}

	meteoClient := meteo.NewOpenMeteoClient(&meteo.OpenMeteoDependencies{})

	crawlerService := crawler.NewService(&crawler.ServiceDependencies{
		DB:          conManager,
		Logger:      log,
		MeteoClient: meteoClient,
	})

	rootCtx := context.Background()
	defer rootCtx.Done()

	if err := crawlerService.Crawl(rootCtx); err != nil {
		log.Error("failed to crawl", err)
	}
}
