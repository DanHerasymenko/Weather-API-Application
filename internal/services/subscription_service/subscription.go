package subscription_service

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/http"
)

type SubscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
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
func (s *SubscriptionService) Subscribe(ctx context.Context, subscriptionRequest *model.SubscriptionCreate) error {

	confirmed, err := s.repo.CheckConfirmation(ctx, subscriptionRequest)

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

// ConfirmSubscription verifies a subscription token and marks the subscription as confirmed.
//
// Returns:
//   - 200 OK if confirmation successful
//   - 400 if already confirmed
//   - 404 if token not found
//   - 500 for unexpected errors
func (s *SubscriptionService) ConfirmSubscription(ctx context.Context, token string) (int, error) {

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
func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) (int, error) {

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

// createNewSubscription inserts a new subscription into the database.
func (s *SubscriptionService) createNewSubscription(ctx context.Context, newToken, email, city, frequency string) error {

	_, err := s.clnts.PostgresClnt.Postgres.Exec(ctx, `INSERT INTO weather_subscriptions (email, city, token, frequency, created_at) VALUES ($1, $2, $3, $4, now())`, email, city, newToken, frequency)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

type Subscription struct {
	Email     string
	City      string
	Frequency string
}

// fetchConfirmedSubscriptions retrieves all confirmed email-city-frequency subscriptions
// from the database to be used for scheduling update routines.
func (s *SubscriptionService) fetchConfirmedSubscriptions(ctx context.Context) ([]Subscription, error) {

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
