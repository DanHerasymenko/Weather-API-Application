package subscription

import (
	"Weather-API-Application/internal/clients"
	"Weather-API-Application/internal/config"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
// Returns:
//   - 200 OK if successfully unsubscribed
//   - 404 if token not found
//   - 500 for DB errors
func (s *Service) Unsubscribe(ctx context.Context, token string) (int, error) {

	result, err := s.clnts.PostgresClnt.Postgres.Exec(ctx, "DELETE FROM weather_subscriptions WHERE token = $1", token)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return http.StatusNotFound, fmt.Errorf("subscription not found")
	}

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
		if err := s.clnts.EmailClnt.SendEmail(email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, newToken)); err != nil {
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
		if err := s.clnts.EmailClnt.SendEmail(email, config.ConfirmSubject, config.BuildConfirmBody(s.cfg.BaseURL, newToken)); err != nil {
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
