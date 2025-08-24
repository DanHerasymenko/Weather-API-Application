package repository

import (
	"Weather-API-Application/internal/model"
	"context"
)

type WeatherRepository interface {
}

type SubscriptionRepository interface {
	Subscribe(ctx context.Context, subscriptionRequest *model.Subscription)
	CheckConfirmation(ctx context.Context, subscriptionRequest *model.Subscription) (rowExists bool, confirmed bool, err error)
	Create(ctx context.Context, subscriptionRequest *model.Subscription) error
	UpdateTokenByEmailCity(ctx context.Context, subscriptionRequest *model.Subscription) error
}
