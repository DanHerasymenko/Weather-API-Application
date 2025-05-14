package weather

import (
	"Weather-API-Application/internal/config"
	"encoding/json"
	"fmt"
	"net/http"
)

type Service struct {
	cfg *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) FetchWeatherForCity(city string) (*Weather, error, int) {

	if s.cfg.WeatherApiKey == "" {
		return nil, fmt.Errorf("weather API key is missing in config"), http.StatusInternalServerError
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", s.cfg.WeatherApiKey, city)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("invalid request: failed to fetch weather data: %w", err), http.StatusBadRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("city not found: failed to fetch weather data: %s", resp.Status), http.StatusNotFound
	}

	var weatherApiResp WeatherAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&weatherApiResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %w", err), http.StatusInternalServerError
	}

	return &Weather{
		Temperature: weatherApiResp.Current.TempC,
		Humidity:    weatherApiResp.Current.Humidity,
		Description: weatherApiResp.Current.Condition.Text,
	}, nil, http.StatusOK
}
