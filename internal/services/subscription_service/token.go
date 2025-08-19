package subscription_service

import (
	"fmt"
	"github.com/google/uuid"
)

func CreateNewToken() string {
	return uuid.New().String()
}

// updateToken updates the token and resets created_at for an existing unconfirmed subscription.
func (s *Service) updateToken(ctx context.Context, token, email, city string) error {

	_, err := s.clnts.PostgresClnt.Postgres.Exec(ctx, "UPDATE weather_subscriptions SET token = $1, created_at = now() WHERE email = $2 AND city = $3", token, email, city)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}
