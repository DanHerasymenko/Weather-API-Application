package subscription

import (
	"Weather-API-Application/internal/logger"
	weather "Weather-API-Application/internal/services/weather"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Scheduler is responsible for spawning and managing weather update routines
// for all confirmed subscriptions.

type Subscription struct {
	Email     string
	City      string
	Frequency string
}

// Start launches background routines for confirmed subscriptions
func (s *Service) StartScheduler(ctx context.Context) error {

	subs, err := s.fetchConfirmedSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch confirmed subscriptions: %w", err)
	}

	for _, sub := range subs {
		go s.startRoutine(ctx, sub)
	}

	logger.Info(ctx, fmt.Sprintf("Starting %d subscription routines", len(subs)))

	return nil
}

func (s *Service) fetchConfirmedSubscriptions(ctx context.Context) ([]Subscription, error) {

	rows, err := s.clnts.PostgresClnt.Postgres.Query(ctx, "SELECT email, city, frequency FROM weather_subscriptions WHERE confirmed = true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		// Scan the row into a Subscription struct
		var sub Subscription
		if err := rows.Scan(&sub.Email, &sub.City, &sub.Frequency); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (s *Service) startRoutine(ctx context.Context, sub Subscription) {

	// Set the interval based on the subscription frequency
	interval := time.Hour
	if strings.ToLower(sub.Frequency) == "daily" {
		interval = 24 * time.Hour
		// Wait until next 8AM (or custom hour)
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), s.cfg.DailyStartHour, 0, 0, 0, now.Location())
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}
		time.Sleep(time.Until(next))
	}

	// Start the ticker for the subscription
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, fmt.Sprintf("Stopping routine for %s - %s", sub.Email, sub.City))
			return
		case <-ticker.C:
			err := s.sendUpdate(sub)
			if err != nil {
				logger.Error(ctx, err)
			}
		}
	}
}

func (s *Service) sendUpdate(sub Subscription) error {

	// Fetch the weather data for the city
	if s.cfg.WeatherApiKey == "" {
		return fmt.Errorf("weather API key is missing in config")
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", s.cfg.WeatherApiKey, sub.City)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("invalid request: failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("city not found: failed to fetch weather data: %s", resp.Status)
	}

	// Decode the response
	var weatherApiResp weather.WeatherAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&weatherApiResp); err != nil {
		return fmt.Errorf("failed to decode weather data: %w", err)
	}

	// Prepare the email content
	weatherMailText := fmt.Sprintf("Weather for %s:\n- temperature: %f\n- humidity: %f\n- description: %s", sub.City, weatherApiResp.Current.TempC, weatherApiResp.Current.Humidity, weatherApiResp.Current.Condition.Text)
	subject := fmt.Sprintf("%s forecast", sub.City)

	// Send the email
	if err := s.clnts.EmailClnt.SendEmail(sub.Email, subject, weatherMailText); err != nil {
		return fmt.Errorf("failed to send email to %s for city %s: %w", sub.Email, sub.City, err)
	}
	return nil
}
