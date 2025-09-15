package repository

import (
	"Weather-API-Application/internal/model"
	"context"
)

type SubscriptionRepository interface {
	CheckConfirmation(ctx context.Context, subscriptionRequest *model.Subscription) (rowExists bool, confirmed bool, err error)
	Create(ctx context.Context, subscriptionRequest *model.Subscription) error
	UpdateTokenByEmailCity(ctx context.Context, subscriptionRequest *model.Subscription) error
	GetByToken(ctx context.Context, token string) (string, *model.Subscription, error)
	SetConfirmed(ctx context.Context, subId string) error
	DeleteByToken(ctx context.Context, token string) error
	ListConfirmed(ctx context.Context) ([]*model.Subscription, error)
}
