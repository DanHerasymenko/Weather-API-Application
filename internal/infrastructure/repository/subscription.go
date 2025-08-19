package repository

import (
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/repository"
	"context"
	"database/sql"
	"errors"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) repository.SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Subscribe(ctx context.Context, subscriptionRequest *model.Subscription) {

}

func (r *SubscriptionRepository) CheckConfirmation(ctx context.Context, subscriptionRequest *model.Subscription) (rowExists bool, confirmed bool, err error) {
	query := `SELECT confirmed FROM weather_subscriptions WHERE email = $1 AND city = $2`
	row := r.db.QueryRowContext(ctx, query, subscriptionRequest.Email, subscriptionRequest.City)
	err = row.Scan(&confirmed)

	if errors.Is(err, sql.ErrNoRows) {
		rowExists = false
		err = nil
		return
	}

	if err != nil {
		return
	}

	rowExists = true
	return
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscriptionRequest *model.Subscription) error {

	//TODO implement me
	panic("implement me")
}
