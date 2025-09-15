package weather_service

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/model"
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

// FetchWeatherForCity retrieves the current weather data for the given city
// using the external WeatherAPI.com services.
//
// It performs the following steps:
//   - Validates that the API key is present in the config.
//   - Sends an HTTP GET request to the Weather API with the city query.
//   - Returns 400 if the request fails to reach the API,
//     404 if the city is not found,
//     or 500 if decoding the response fails.
//   - On success, returns a populated Weather struct with temperature, humidity, and description.
func (s *Service) FetchWeatherForCity(city string) (*model.Weather, error, int) {

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

	var weatherApiResp model.WeatherAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&weatherApiResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %w", err), http.StatusInternalServerError
	}

	return &model.Weather{
		Temperature: weatherApiResp.Current.TempC,
		Humidity:    weatherApiResp.Current.Humidity,
		Description: weatherApiResp.Current.Condition.Text,
	}, nil, http.StatusOK
}
