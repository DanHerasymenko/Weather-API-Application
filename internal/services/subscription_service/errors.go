package subscription_service

import "errors"

var (
	ErrSubscriptionExists         = errors.New("subscription already exists")
	ErrNotFound                   = errors.New("subscription not found")
	ErrAlreadyConfirmed           = errors.New("subscription already confirmed")
	ErrFailedToCreateSubscription = errors.New("failed to create subscription")
)
