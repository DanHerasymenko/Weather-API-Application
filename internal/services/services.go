package services

import (
	"Weather-API-Application/internal/clients"
	"Weather-API-Application/internal/config"
)

type Services struct {
}

func NewServices(cfg *config.Config, clnts *clients.Clients) *Services {
	return &Services{}
}
