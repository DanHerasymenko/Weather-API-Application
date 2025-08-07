package repository

import (
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/repository"
	"context"
	"database/sql"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) repository.SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Subscribe(ctx context.Context, subscriptionRequest *model.SubscriptionCreate) {

}

func (r *SubscriptionRepository) CheckConfirmation(ctx context.Context, subscriptionRequest *model.SubscriptionCreate) (bool, error) {
	var confirmed bool
	query := `SELECT confirmed FROM weather_subscriptions WHERE email = $1 AND city = $2`

	row := r.db.QueryRowContext(ctx, query, subscriptionRequest.Email, subscriptionRequest.City)
	err := row.Scan(&confirmed)

	if err != nil {
		return confirmed, err
	}
	return confirmed, nil
}
