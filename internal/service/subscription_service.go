package service

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ValidateCity checks if the provided city is valid by making a request to the WeatherAPI.com service.
func (s *Service) ValidateCity(city string) (bool, error, int) {
	if s.cfg.WeatherApiKey == "" {
		return false, fmt.Errorf("weather API key is missing in config"), http.StatusInternalServerError
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", s.cfg.WeatherApiKey, city)

	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to fetch weather data: %w", err), http.StatusBadRequest
	}
	defer resp.Body.Close()

	var errResp struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != nil {
		return false, fmt.Errorf("city not found: %s", errResp.Error.Message), http.StatusBadRequest
	}

	return true, nil, http.StatusOK
}

// Scheduler is responsible for spawning and managing weather update routines
// for all confirmed subscriptions.

type Subscription struct {
	Email     string
	City      string
	Frequency string
}

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

// fetchConfirmedSubscriptions retrieves all confirmed email-city-frequency subscriptions
// from the database to be used for scheduling update routines.
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

func (s *Service) MakeKey(sub Subscription) string {
	return fmt.Sprintf("%s|%s", sub.Email, strings.ToLower(sub.City))
}

type Service struct {
	cfg   *config.Config
	clnts *client.Clients

	routines map[string]context.CancelFunc
	mu       sync.Mutex
}

func NewService(cfg *config.Config, clnts *client.Clients) *Service {
	return &Service{
		cfg:      cfg,
		clnts:    clnts,
		routines: make(map[string]context.CancelFunc),
	}
}

// ConfirmSubscription verifies a subscription token and marks the subscription as confirmed.
//
// Returns:
//   - 200 OK if confirmation successful
//   - 400 if already confirmed
//   - 404 if token not found
//   - 500 for unexpected errors
func (s *Service) ConfirmSubscription(ctx context.Context, token string) (int, error) {

	row := s.clnts.PostgresClnt.Postgres.QueryRow(ctx, "SELECT id, email, city, frequency, confirmed FROM weather_subscriptions WHERE token = $1", token)

	var id int
	var sub Subscription
	var confirmed bool

	err := row.Scan(&id, &sub.Email, &sub.City, &sub.Frequency, &confirmed)
	if errors.Is(err, pgx.ErrNoRows) {
		return http.StatusNotFound, fmt.Errorf("subscription not found")
	}
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to scan subscription: %w", err)
	}
	if confirmed {
		return http.StatusBadRequest, fmt.Errorf("subscription already confirmed")
	}

	// Mark subscription as confirmed
	_, err = s.clnts.PostgresClnt.Postgres.Exec(ctx, "UPDATE weather_subscriptions SET confirmed = true WHERE id = $1", id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update subscription: %w", err)
	}

	// Start the weather update routine for the confirmed subscription
	go s.startRoutine(ctx, sub)

	return http.StatusOK, nil
}

// Unsubscribe removes a weather subscription using the provided confirmation token.
//
// This method performs three steps:
//  1. Retrieves the subscription (email, city, frequency) by token.
//  2. Deletes the subscription from the database.
//  3. Stops the background weather update routine associated with that subscription,
//     by calling the stored context.CancelFunc and removing it from the routines map.
//
// Returns:
//   - 200 OK if successfully unsubscribed and routine was stopped.
//   - 404 if the token does not match any subscription.
//   - 500 if a database or internal error occurred.
func (s *Service) Unsubscribe(ctx context.Context, token string) (int, error) {

	// 1.Fetch the subscription details using the token
	var sub Subscription
	query := "SELECT email, city, frequency FROM weather_subscriptions WHERE token = $1"
	err := s.clnts.PostgresClnt.Postgres.QueryRow(ctx, query, token).
		Scan(&sub.Email, &sub.City, &sub.Frequency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return http.StatusNotFound, fmt.Errorf("subscription not found")
		}
		return http.StatusInternalServerError, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// 2.Delete the subscription from the database
	_, err = s.clnts.PostgresClnt.Postgres.Exec(ctx, "DELETE FROM weather_subscriptions WHERE token = $1", token)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete subscription: %w", err)
	}

	// 3.Stop the routine
	key := s.MakeKey(sub)

	s.mu.Lock()
	if cancel, ok := s.routines[key]; ok {
		cancel()
		delete(s.routines, key)
	}
	s.mu.Unlock()

	return http.StatusOK, nil
}

// Subscribe attempts to create or update a weather subscription for a given email and city.
//
// The logic flow is as follows:
//  1. Checks if a subscription for the email + city pair already exists.
//  2. If found and confirmed = true → returns 409 Conflict.
//  3. If found and confirmed = false → generates a new token, updates the record, and sends a new confirmation email.
//  4. If not found → creates a new record with a generated token and sends confirmation email.
//
// Returns:
//   - 200 OK on successful creation or update.
//   - 409 Conflict if subscription is already active.
//   - 500 Internal Server Error on DB or email errors.
func (s *Service) Subscribe(ctx context.Context, email, city, frequency string) (int, error) {

	row := s.clnts.PostgresClnt.Postgres.QueryRow(ctx, "SELECT confirmed FROM weather_subscriptions WHERE email = $1 AND city = $2", email, city)

	var confirmed bool

	err := row.Scan(&confirmed)

	//found and confirmed - true
	if err == nil && confirmed {
		return http.StatusConflict, fmt.Errorf("subscription already exists")
	}

	//found and confirmed - false
	if err == nil && !confirmed {

		// update token
		newToken := createNewToken()
		err := s.updateToken(ctx, newToken, email, city)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to update subscription: %w", err)
		}

		// send confirmation email
		if err := s.clnts.EmailClnt.SendEmail(ctx, email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, newToken)); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to send confirmation email: %w", err)
		}
		return http.StatusOK, nil
	}

	//not found - create new sub
	if errors.Is(err, pgx.ErrNoRows) {

		// create new token
		newToken := createNewToken()
		err := s.createNewSubscription(ctx, newToken, email, city, frequency)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to create subscription: %w", err)
		}

		// send confirmation email
		if err := s.clnts.EmailClnt.SendEmail(ctx, email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, newToken)); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to send confirmation email: %w", err)
		}
		return http.StatusOK, nil
	}
	return http.StatusOK, nil
}

func createNewToken() string {
	return uuid.New().String()
}

// updateToken updates the token and resets created_at for an existing unconfirmed subscription.
func (s *Service) updateToken(ctx context.Context, token, email, city string) error {

	_, err := s.clnts.PostgresClnt.Postgres.Exec(ctx, "UPDATE weather_subscriptions SET token = $1, created_at = now() WHERE email = $2 AND city = $3", token, email, city)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// createNewSubscription inserts a new subscription into the database.
func (s *Service) createNewSubscription(ctx context.Context, newToken, email, city, frequency string) error {

	_, err := s.clnts.PostgresClnt.Postgres.Exec(ctx, `INSERT INTO weather_subscriptions (email, city, token, frequency, created_at) VALUES ($1, $2, $3, $4, now())`, email, city, newToken, frequency)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}
