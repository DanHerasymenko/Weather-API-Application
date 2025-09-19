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

var ErrNotFound = errors.New("subscription not found")

func NewSubscriptionRepository(db *sql.DB) repository.SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) CheckConfirmation(ctx context.Context, subscriptionRequest *model.Subscription) (rowExists bool, confirmed bool, err error) {
	const query = `
		SELECT confirmed
		FROM weather_subscriptions
		WHERE email = $1 AND city = $2
	`
	row := r.db.QueryRowContext(ctx, query, subscriptionRequest.Email, subscriptionRequest.City)
	err = row.Scan(&confirmed)

	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return true, confirmed, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *model.Subscription) error {
	const query = `
		INSERT INTO weather_subscriptions (email, city, token, frequency, confirmed, created_at)
		VALUES ($1,   $2,   $3,    $4,       FALSE,     NOW())
	`
	_, err := r.db.ExecContext(ctx, query, s.Email, s.City, s.Token, s.Frequency)
	return err
}

func (r *SubscriptionRepository) UpdateTokenByEmailCity(ctx context.Context, s *model.Subscription) error {
	const query = `
		UPDATE weather_subscriptions
		SET token = $1, confirmed = FALSE, created_at = NOW()
		WHERE email = $2 AND city = $3
	`
	res, err := r.db.ExecContext(ctx, query, s.Token, s.Email, s.City)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		// No rows affected - return domain error
		return ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepository) GetByToken(ctx context.Context, token string) (string, *model.Subscription, error) {
	const query = `
		SELECT id, email, city, frequency, confirmed
		FROM weather_subscriptions
		WHERE token = $1
	`
	var (
		id        string
		email     string
		city      string
		frequency string
		confirmed bool
	)
	err := r.db.QueryRowContext(ctx, query, token).Scan(&id, &email, &city, &frequency, &confirmed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, ErrNotFound
	}
	if err != nil {
		return "", nil, err
	}

	return id, &model.Subscription{
		Email:     email,
		City:      city,
		Frequency: frequency,
		Token:     token,
		Confirmed: confirmed,
	}, nil
}

func (r *SubscriptionRepository) SetConfirmed(ctx context.Context, subId string) error {
	const query = `
		UPDATE weather_subscriptions
		SET confirmed = TRUE, confirmed_at = NOW()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, subId)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepository) DeleteByToken(ctx context.Context, token string) error {
	const query = `
		DELETE FROM weather_subscriptions
		WHERE token = $1
	`
	res, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepository) ListConfirmed(ctx context.Context) ([]*model.Subscription, error) {
	const query = `
		SELECT email, city, frequency, token, confirmed
		FROM weather_subscriptions
		WHERE confirmed = TRUE
		ORDER BY email, city
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*model.Subscription
	for rows.Next() {
		s := new(model.Subscription)
		if err := rows.Scan(&s.Email, &s.City, &s.Frequency, &s.Token, &s.Confirmed); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}
