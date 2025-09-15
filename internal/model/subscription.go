package model

type Subscription struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
	Token     string `json:"token"`
	Confirmed bool   `json:"confirmed"`
}
