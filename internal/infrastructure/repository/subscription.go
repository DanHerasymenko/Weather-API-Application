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
	query := `INSERT INTO weather_subscriptions (email, city, token, frequency, created_at) VALUES ($1, $2, $3, $4, now())`

	return r.db.QueryRowContext(
		ctx, query,
		subscriptionRequest.Email,
		subscriptionRequest.City,
		subscriptionRequest.Token,
		subscriptionRequest.Frequency)
}

func (r *SubscriptionRepository) UpdateTokenByEmailCity(ctx context.Context, subscriptionRequest *model.Subscription) error {
	query := `UPDATE weather_subscriptions SET token = $1, created_at = now() WHERE email = $2 AND city = $3`

	return r.db.QueryRowContext(
		ctx, query,
		subscriptionRequest.Token,
		subscriptionRequest.Email,
		subscriptionRequest.City)
}

func (r *SubscriptionRepository) GetByToken(ctx context.Context, token string) (string, *model.Subscription, error) {
	query := `SELECT email, city, frequency FROM weather_subscriptions WHERE token = $1`

	row := r.db.QueryRowContext(ctx, token)
	err := row.Scan(&row)
	//if err != nil {
	//	return model.Subscription{
	//		Email:     row.,
	//		City:      "",
	//		Frequency: "",
	//		Token:     "",
	//		Confirmed: false,
	//	}
	//}
}

func (r *SubscriptionRepository) SetConfirmed(ctx context.Context, subId string) error {
	return nil
}

func (r *SubscriptionRepository) DeleteByToken(ctx context.Context, token string) error {
	return nil
}

func (r *SubscriptionRepository) ListConfirmed(ctx context.Context) ([]*model.Subscription, error) {
	return nil, nil
}
