package crawler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"temperature-checker/internal/core/meteo"
	"temperature-checker/internal/db"
	dbGen "temperature-checker/internal/db/gen"
	"time"
)

type ServiceDependencies struct {
	DB          *db.ConManager
	Logger      *slog.Logger
	MeteoClient meteo.Client
}

type Service struct {
	db *db.ConManager
	l  *slog.Logger
	mc meteo.Client
}

func NewService(deps *ServiceDependencies) *Service {
	return &Service{
		db: deps.DB,
		l:  deps.Logger,
		mc: deps.MeteoClient,
	}
}

func (s *Service) Crawl(ctx context.Context) error {
	q := s.db.WithQ()

	locations, err := q.GetAPILocationSensors(ctx)

	if err != nil {
		return fmt.Errorf("get locations: %w", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(locations)*2)

	for _, l := range locations {
		wg.Add(1)
		go func(l dbGen.GetAPILocationSensorsRow) {
			defer func(errCh chan error) {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("panic in pullWeatherUpdate for %v: %v", l, r)
				}
			}(errCh)

			defer wg.Done()

			if err := s.pullWeatherUpdate(ctx, l); err != nil {
				errCh <- fmt.Errorf("location %v: %w", l, err)
			}
		}(l)
	}

	wg.Wait()
	close(errCh)

	var allErr error
	for err := range errCh {
		allErr = errors.Join(allErr, err)
	}

	return allErr
}

func (s *Service) pullWeatherUpdate(ctx context.Context, l dbGen.GetAPILocationSensorsRow) error {
	res, err := s.mc.GetWeather(ctx, meteo.WeatherParams{
		Lat: l.Latitude,
		Lon: l.Longitude,
	})
	if err != nil {
		return fmt.Errorf("get weather: %w", err)
	}

	data := s.processResponse(l.LocationSensorID, res)

	if _, err := s.db.WithQ().CreateTemperatureData(ctx, data); err != nil {
		return fmt.Errorf("create temperature data: %w", err)
	}

	s.l.Info("weather data for location saved", "locationName", l.LocationName, "sensor", l.SensorID)

	return nil
}

func (s *Service) processResponse(lsId int32, res []meteo.WeatherData) dbGen.CreateTemperatureDataParams {

	n := len(res)

	sensorIds := make([]int32, n)
	values := make([]float64, n)
	timestamps := make([]time.Time, n)

	for i, r := range res {
		sensorIds[i] = lsId
		values[i] = r.Temperature
		timestamps[i] = r.Timestamp
	}

	return dbGen.CreateTemperatureDataParams{
		LocationSensorIds: sensorIds,
		Values:            values,
		Timestamps:        timestamps,
	}
}
