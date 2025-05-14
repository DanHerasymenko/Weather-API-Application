package subscription

import (
	"Weather-API-Application/internal/clients"
	"Weather-API-Application/internal/config"
)

type Service struct {
	cfg   *config.Config
	clnts *clients.Clients
}
