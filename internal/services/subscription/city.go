package subscription

import (
	"fmt"
	"net/http"
)

// ValidateCity checks if the provided city is valid by making a request to the WeatherAPI.com service.
func (s *Service) ValidateCity(city string) (bool, error, int) {

	if s.cfg.WeatherApiKey == "" {
		return false, fmt.Errorf("can not validate city: weather API key is missing in config"), http.StatusInternalServerError
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", s.cfg.WeatherApiKey, city)

	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("invalid request, failed to validate city: failed to fetch weather data: %w", err), http.StatusBadRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("city not found: failed to fetch weather data: %s", resp.Status), http.StatusNotFound
	}

	return true, nil, http.StatusOK
}
