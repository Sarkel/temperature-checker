package meteo

import (
	"context"
	"time"
)

type WeatherParams struct {
	Lat float64
	Lon float64
}

type WeatherData struct {
	Timestamp   time.Time
	Temperature float64
}

type Client interface {
	GetWeather(ctx context.Context, params WeatherParams) ([]WeatherData, error)
}
