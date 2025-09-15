package handler

import (
	"fmt"
	"net/http"

	"Weather-API-Application/internal/services/weather_service"
	"Weather-API-Application/internal/utils/response"
	"Weather-API-Application/internal/utils/validate"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	svc weather_service.WeatherService
}

func NewWeatherHandler(svc weather_service.WeatherService) *WeatherHandler {
	return &WeatherHandler{svc: svc}
}

// RegisterRoutes registers weather endpoints.
func (h *WeatherHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.GET("/weather", h.GetWeather)
	}
}

// GetWeather godoc
// @Summary      Get current weather for a city
// @Description  Returns the current weather for the specified city using WeatherAPI.com.
// @Tags         weather
// @Accept       json
// @Produce      json
// @Param        city  query     string  true  "City name"
// @Success      200   {object}  model.Weather  "Current weather returned"
// @Failure      400   {object}  response.ErrorResponse   "Invalid request"
// @Failure      404   {object}  response.ErrorResponse   "City not found"
// @Router       /weather [get]
func (h *WeatherHandler) GetWeather(ctx *gin.Context) {
	city := ctx.Query("city")

	// Validate input
	if !validate.IsValidCity(city) {
		response.AbortWithErrorJSON(ctx, http.StatusBadRequest,
			ctx.Error(fmt.Errorf("invalid city parameter")),
			"City parameter is required and cannot be empty")
		return
	}

	fetchedWeather, err, code := h.svc.FetchWeatherForCity(city)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}
	ctx.JSON(200, fetchedWeather)
}
