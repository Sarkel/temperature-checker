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
	"temperature-checker/internal/mqtt"
	"time"
)

type ServiceDependencies struct {
	DB          *db.ConManager
	Logger      *slog.Logger
	MeteoClient meteo.Client
	Broker      mqtt.Client
}

type Service struct {
	db *db.ConManager
	l  *slog.Logger
	mc meteo.Client
	b  mqtt.Client
}

func NewService(deps *ServiceDependencies) *Service {
	return &Service{
		db: deps.DB,
		l:  deps.Logger,
		mc: deps.MeteoClient,
		b:  deps.Broker,
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

	// todo: create topic utilities
	topic := fmt.Sprintf("sensors/%s/%s", l.LocationSid, l.SensorSid)

	data := s.processResponse(res)

	if err := s.b.Publish(topic, data); err != nil {
		return fmt.Errorf("publish temperature data: %w", err)
	}

	s.l.Info("weather data for location saved", "locationName", l.LocationName, "sensor", l.SensorSid)

	return nil
}

func (s *Service) processResponse(res []meteo.WeatherData) []mqtt.MessagePayload {
	n := len(res)

	results := make([]mqtt.MessagePayload, n)

	for i, r := range res {
		results[i] = []string{fmt.Sprintf("%.2f", r.Temperature), r.Timestamp.Format(time.RFC3339)}
	}

	return results
}
