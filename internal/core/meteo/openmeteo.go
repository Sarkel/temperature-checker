package meteo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const openMeteoTimeLayout = "2006-01-02T15:04"

type OpenMeteoDependencies struct{}
type OpenMeteoClient struct{}

func NewOpenMeteoClient(_ *OpenMeteoDependencies) *OpenMeteoClient {
	return &OpenMeteoClient{}
}

func (s *OpenMeteoClient) GetWeather(ctx context.Context, params WeatherParams) ([]WeatherData, error) {
	url := s.buildUrl(params.Lat, params.Lon)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request canceled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to get weather from openmeteo: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from openmeteo: %s", resp.Status)
	}

	var data OpenMeteoResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	res, err := s.mapResponse(data)

	if err != nil {
		return nil, fmt.Errorf("map response: %w", err)
	}

	return res, nil
}

func (s *OpenMeteoClient) buildUrl(lat float64, lon float64) string {
	var b strings.Builder

	b.WriteString("https://api.open-meteo.com/v1/forecast")
	b.WriteString("?current_weather=true")
	b.WriteString(fmt.Sprintf("&latitude=%f", lat))
	b.WriteString(fmt.Sprintf("&longitude=%f", lon))

	return b.String()
}

func (s *OpenMeteoClient) mapResponse(resp OpenMeteoResponse) ([]WeatherData, error) {
	t, err := time.Parse(openMeteoTimeLayout, resp.CurrentWeather.Time)
	if err != nil {
		return nil, fmt.Errorf("parse current weather time: %w", err)
	}

	return []WeatherData{{
		Temperature: resp.CurrentWeather.Temperature,
		Timestamp:   t,
	}}, nil
}
