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

func (s *Service) ConfirmSubscription(ctx context.Context, token string) (int, error) {

	row := s.clnts.PostgresClnt.Postgres.QueryRow(ctx, "SELECT id, confirmed FROM weather_subscriptions WHERE token = $1", token)

	var id uuid.UUID
	var confirmed bool

	err := row.Scan(&id, &confirmed)
	if errors.Is(err, pgx.ErrNoRows) {
		return http.StatusNotFound, fmt.Errorf("subscription not found")
	}

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to scan subscription: %w", err)
	}

	if confirmed {
		return http.StatusBadRequest, fmt.Errorf("subscription already confirmed")
	}

	_, err = s.clnts.PostgresClnt.Postgres.Exec(ctx, "UPDATE weather_subscriptions SET confirmed = true WHERE id = $1", id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update subscription: %w", err)
	}

	return http.StatusOK, nil
}

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

func (s *Service) Subscribe(ctx context.Context, email, city string) (int, error) {

	row := s.clnts.PostgresClnt.Postgres.QueryRow(ctx, "SELECT confirmed FROM weather_subscriptions WHERE email = $1 AND city = $2", email, city)

	var confirmed bool

	err := row.Scan(&confirmed)

	//found and confirmed - true
	if err == nil && confirmed {
		return http.StatusConflict, fmt.Errorf("subscription already exists")
	}

	//found and confirmed - false
	if err == nil && !confirmed {
		err := s.updateToken(ctx, email, city)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to update subscription: %w", err)
		}
		// send email
	}

	//not found - create new sub
	if errors.Is(err, pgx.ErrNoRows) {
		err := s.createNewSubscription(ctx, email, city)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to create subscription: %w", err)
		}
		// send email
	}

	return 200, nil
}

func createNewToken() string {
	return uuid.New().String()
}

func (s *Service) updateToken(ctx context.Context, email, city string) error {

	_, err := s.clnts.PostgresClnt.Postgres.Exec(ctx, "UPDATE weather_subscriptions SET token = $1 WHERE email = $1 AND city = $3", createNewToken(), email, city)
	// TODO: HOW TO UPDATE CREATED AT
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *Service) createNewSubscription(ctx context.Context, email, city string) error {
	_, err := s.clnts.PostgresClnt.Postgres.Exec(ctx, "INSERT INTO weather_subscriptions (email, city, token) VALUES ($1, $2, $3)", email, city, createNewToken())
	// TODO: HOW TO UPDATE CREATED AT
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func sendEmail() {
	// send email
	// use smtp client
	// use html template
	// use email service
}
