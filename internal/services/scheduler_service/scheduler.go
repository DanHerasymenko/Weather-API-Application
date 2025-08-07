package scheduler_service

import (
	"Weather-API-Application/internal/logger"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Scheduler is responsible for spawning and managing weather update routines
// for all confirmed subscriptions.

// StartScheduler fetches all confirmed subscriptions from the database
// and starts a background goroutine for each one to send periodic weather updates.
func (s *Service) StartScheduler(ctx context.Context) error {

	subs, err := s.fetchConfirmedSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch confirmed subscriptions: %w", err)
	}

	for _, sub := range subs {

		// For each subscription, create a new own context with a cancel function
		subCtx, cancel := context.WithCancel(ctx)
		key := s.MakeKey(sub)

		s.mu.Lock()
		s.routines[key] = cancel
		s.mu.Unlock()

		// Start the routine for this particular subscription
		go s.startRoutine(subCtx, sub)
	}

	logger.Info(ctx, fmt.Sprintf("Starting %d subscription routines", len(subs)))

	return nil
}

// startRoutine runs a background loop for a single subscription.
// It determines whether the updates should be sent hourly or daily,
// waits for the correct interval, then periodically calls sendUpdate.
// The routine stops when the provided context is cancelled.
func (s *Service) startRoutine(ctx context.Context, sub Subscription) {

	// Detect the interval based on the subscription frequency
	interval := time.Hour
	if strings.ToLower(sub.Frequency) == "daily" {
		interval = 24 * time.Hour

		// Calculate how long to wait until the next daily update
		now := time.Now()
		next := time.Date(
			now.Year(), now.Month(), now.Day(),
			s.cfg.DailyStartHour, 0, 0, 0,
			now.Location(),
		)

		// If current time is after the next scheduled time, add 24 hours
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		// Wait until the next scheduled time
		select {
		case <-time.After(time.Until(next)):
			// continue to the loop with proper start time
		case <-ctx.Done():
			logger.Info(ctx, fmt.Sprintf("Routine cancelled before first run: %s - %s", sub.Email, sub.City))
			return
		}
	} else {
		interval = time.Hour
	}

	// Create a ticker to send updates with the specified interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, fmt.Sprintf("Stopping routine for %s - %s", sub.Email, sub.City))
			return

		case <-ticker.C:

			logger.Info(ctx, fmt.Sprintf("Attempting to send update to %s for %s", sub.Email, sub.City))
			if err := s.sendUpdate(ctx, sub); err != nil {
				logger.Error(ctx, err)
			}
			logger.Info(ctx, fmt.Sprintf("Weather update sent to %s for city %s", sub.Email, sub.City))
		}
	}
}

// sendUpdate fetches the current weather data for the given subscription's city,
// formats it into a plain text message, and sends it via email to the subscriber.
// Returns an error if any of the steps fail (HTTP request, JSON parsing, or email sending).
func (s *Service) sendUpdate(ctx context.Context, sub Subscription) error {

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

	var weatherApiResp weather.WeatherAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&weatherApiResp); err != nil {
		return fmt.Errorf("failed to decode weather data: %w", err)
	}

	// Prepare the email content
	weatherMailText := fmt.Sprintf(`Weather for %s:<br>- temperature: %.1f°C<br>- humidity: %.0f%%<br>- description: %s`,
		sub.City, weatherApiResp.Current.TempC, weatherApiResp.Current.Humidity, weatherApiResp.Current.Condition.Text)
	subject := fmt.Sprintf("%s forecast", sub.City)

	// Send the email
	if err := s.clnts.EmailClnt.SendEmail(ctx, sub.Email, subject, weatherMailText); err != nil {
		return fmt.Errorf("failed to send email to %s for city %s: %w", sub.Email, sub.City, err)
	}

	return nil
}
