package weather

import (
	"Weather-API-Application/internal/clients"
	"Weather-API-Application/internal/config"
	"encoding/json"
	"fmt"
	"net/http"
)

type Service struct {
	cfg   *config.Config
	clnts *clients.Clients
}

func NewService(cfg *config.Config, clnts *clients.Clients) *Service {
	return &Service{
		cfg:   cfg,
		clnts: clnts,
	}
}

func (s *Service) FetchWeatherForCity(city string) (*Weather, error) {

	var weatherApiResp WeatherAPIResponse

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", city, s.cfg.WeatherApiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch weather data: %s", resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&weatherApiResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %w", err)
	}

	return &Weather{
		Temperature: weatherApiResp.Current.TempC,
		Humidity:    weatherApiResp.Current.Humidity,
		Description: weatherApiResp.Current.Condition.Text,
	}, nil
}
