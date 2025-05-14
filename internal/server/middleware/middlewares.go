package middleware

import (
	"Weather-API-Application/internal/clients"
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/server/middleware/logging"
)

type Middlewares struct {
	Log *logging.Middleware
}

func NewMiddlewares(cfg *config.Config, clnts *clients.Clients) *Middlewares {
	return &Middlewares{
		Log: logging.NewMiddleware(cfg),
	}
}
