package handlers

import (
	"Weather-API-Application/internal/config"
	sh "Weather-API-Application/internal/server/handlers/subscription"
	wh "Weather-API-Application/internal/server/handlers/weather"
	"Weather-API-Application/internal/server/middleware"
	"Weather-API-Application/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Weather      *wh.Handler
	Subscription *sh.Handler

	mdlwrs *middleware.Middlewares
}

func NewHandlers(cfg *config.Config, srvc *services.Services, mdlwrs *middleware.Middlewares) *Handlers {
	return &Handlers{
		Weather:      wh.NewHandler(srvc),
		Subscription: sh.NewHandler(cfg, srvc),

		mdlwrs: mdlwrs,
	}
}

// RegisterRoutes sets up the API routes and middleware for the application.
func (h *Handlers) RegisterRoutes(router *gin.Engine) {

	// Swagger + static index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Static("/static", "./static")

	// API group with global log middleware
	api := router.Group("/api")
	api.Use(h.mdlwrs.Log.Handle)

	// Weather GET
	api.GET("/weather", h.Weather.GetWeather)

	// Subscription endpoints
	api.POST("/subscribe", h.Subscription.Subscribe)
	api.GET("/confirm/:token", h.Subscription.ConfirmSubscription)
	api.GET("/unsubscribe/:token", h.Subscription.Unsubscribe)
}
