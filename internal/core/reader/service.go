package reader

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"temperature-checker/internal/db"
	genDb "temperature-checker/internal/db/gen"
	"temperature-checker/internal/mqtt"
	"time"
)

type Dependencies struct {
	DB     *db.ConManager
	Logger *slog.Logger
	Broker mqtt.Client
}

type Service struct {
	db *db.ConManager
	l  *slog.Logger
	b  mqtt.Client
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		db: deps.DB,
		l:  deps.Logger,
		b:  deps.Broker,
	}
}

func (s *Service) Listen(ctx context.Context) error {
	if err := s.b.Subscribe(ctx, "sensors/#", s.processMessage); err != nil {
		return fmt.Errorf("temp reader subscribe: %w", err)
	}
	return nil
}

func (s *Service) processMessage(ctx context.Context, _ mqtt.Client, msg mqtt.Message) {
	// todo: move logic to save to DB to separate goroutine with queue process
	q := s.db.WithQ()

	locationSensorId, err := s.getLocationSensorId(ctx, &msg)

	if err != nil {
		s.l.Error("failed to get location sensor id", "err", err)
	}

	sensorData, err := s.parseSensorData(locationSensorId, &msg)

	if err != nil {
		s.l.Error("failed to parse sensor data", "err", err)
	}

	if _, err := q.CreateTemperatureData(ctx, sensorData); err != nil {
		s.l.Error("failed to save temperature data", "err", err)
	}
	s.l.Info("temperature data saved", "topic", msg.Topic)
}

func (s *Service) getLocationSensorId(ctx context.Context, msg *mqtt.Message) (int32, error) {
	parts := strings.Split(msg.Topic, "/")

	if len(parts) != 3 {
		return -1, fmt.Errorf("invalid topic format %s", msg.Topic)
	}

	locationSensorId, err := s.db.WithQ().GetLocationSensorBySensorId(ctx, genDb.GetLocationSensorBySensorIdParams{
		SensorSid:   parts[2],
		LocationSid: parts[1],
	})

	if err != nil {
		return -1, fmt.Errorf("failed to get location sensor id %w", err)
	}

	return locationSensorId, nil
}

func (s *Service) parseSensorData(locationSensorId int32, msg *mqtt.Message) (genDb.CreateTemperatureDataParams, error) {
	n := len(msg.Payload)

	locationSensorIds := make([]int32, n)
	sensorValues := make([]float64, n)
	sensorTimes := make([]time.Time, n)

	for i, p := range msg.Payload {
		if len(p) != 2 {
			return genDb.CreateTemperatureDataParams{}, fmt.Errorf("invalid payload length %d", len(p))
		}

		locationSensorIds[i] = locationSensorId

		sensorValue, err := strconv.ParseFloat(p[0], 64)

		if err != nil {
			return genDb.CreateTemperatureDataParams{}, fmt.Errorf("failed to parse sensor value %w", err)
		}

		sensorValues[i] = sensorValue

		sensorTime, err := time.Parse(time.RFC3339, p[1])

		if err != nil {
			return genDb.CreateTemperatureDataParams{}, fmt.Errorf("failed to parse sensor time %w", err)
		}

		sensorTimes[i] = sensorTime
	}

	return genDb.CreateTemperatureDataParams{
		LocationSensorIds: locationSensorIds,
		Values:            sensorValues,
		Timestamps:        sensorTimes,
	}, nil
}
