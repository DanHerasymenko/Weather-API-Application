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

func (h *Handlers) RegisterRoutes(router *gin.Engine) {

	// Swagger + static index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Static("/static", "./static")

	// API group with global log middleware
	api := router.Group("/api")
	api.Use(h.mdlwrs.Log.Handle)

	// Weather GET
	weather := api.Group("/weather")
	weather.GET("/", h.Weather.GetWeather)

	// Subscription operations
	subscription := api.Group("/subscription")
	subscription.POST("/subscribe", h.Subscription.Subscribe)
	subscription.GET("/confirm/:token", h.Subscription.ConfirmSubscription)
}
