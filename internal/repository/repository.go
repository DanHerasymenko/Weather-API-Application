package repository

import (
	"Weather-API-Application/internal/model"
	"context"
)

type WeatherRepository interface {
}

type SubscriptionRepository interface {
	Subscribe(ctx context.Context, subscriptionRequest *model.SubscriptionCreate)
	CheckConfirmation(ctx context.Context, subscriptionRequest *model.SubscriptionCreate) (bool, error)
}
