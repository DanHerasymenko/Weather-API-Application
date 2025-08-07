package model

type SubscriptionCreate struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
}
