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

	if len(msg.Payload) != 2 {
		s.l.Error("invalid payload length", "payload", msg.Payload)
		return
	}

	locationSensorId, err := s.getLocationSensorId(ctx, &msg)

	if err != nil {
		s.l.Error("failed to get location sensor id", "err", err)
	}

	sensorValue, sensorTime, err := s.parseSensorData(&msg)

	if err != nil {
		s.l.Error("failed to parse sensor data", "err", err)
	}

	_, err = q.CreateTemperatureData(ctx, genDb.CreateTemperatureDataParams{
		LocationSensorIds: []int32{locationSensorId},
		Values:            []float64{sensorValue},
		Timestamps:        []time.Time{sensorTime},
	})

	if err != nil {
		s.l.Error("failed to save temperature data", "err", err)
	}
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

func (s *Service) parseSensorData(msg *mqtt.Message) (float64, time.Time, error) {
	sensorValue, err := strconv.ParseFloat(msg.Payload[1], 64)

	if err != nil {
		return -1., time.Time{}, fmt.Errorf("failed to parse sensor value %w", err)
	}

	sensorTime, err := time.Parse(time.RFC3339, msg.Payload[2])

	if err != nil {
		return -1., time.Time{}, fmt.Errorf("failed to parse sensor time %w", err)
	}

	return sensorValue, sensorTime, nil
}
