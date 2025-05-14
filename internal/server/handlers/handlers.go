package handlers

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/server/middleware"
	"Weather-API-Application/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	mdlwrs *middleware.Middlewares
}

func NewHandlers(cfg *config.Config, srvc *services.Services, mdlwrs *middleware.Middlewares) *Handlers {
	return &Handlers{
		mdlwrs: mdlwrs,
	}
}

func (h *Handlers) RegisterRoutes(router *gin.Engine) {

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API group with global log middleware
	api := router.Group("/api")
	api.Use(h.mdlwrs.Log.Handle)
}
