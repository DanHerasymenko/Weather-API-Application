package services

import (
	"Weather-API-Application/internal/clients"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/services/weather"
)

type Services struct {
	Weather *weather.Service
}

func NewServices(cfg *config.Config, clnts *clients.Clients) *Services {
	return &Services{
		Weather: weather.NewService(cfg, clnts),
	}
}
