package subscription_service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var errResp struct {
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ValidateCity checks if the provided city is valid by making a request to the WeatherAPI.com services.
func (s *SubscriptionService) ValidateCity(city string) (bool, error, int) {
	if s.cfg.WeatherApiKey == "" {
		return false, fmt.Errorf("weather API key is missing in config"), http.StatusInternalServerError
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", s.cfg.WeatherApiKey, city)

	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to fetch weather data: %w", err), http.StatusBadRequest
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != nil {
		return false, fmt.Errorf("city not found: %s", errResp.Error.Message), http.StatusBadRequest
	}

	return true, nil, http.StatusOK
}
