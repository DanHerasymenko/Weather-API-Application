package weather

import (
	"Weather-API-Application/internal/services"
	"Weather-API-Application/internal/utils/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	srvc *services.Services
}

func NewHandler(srvc *services.Services) *Handler {
	return &Handler{
		srvc: srvc,
	}
}

// GetWeather godoc
// @Summary      Get current weather for a city
// @Description  Returns the current weather forecast for the specified city using WeatherAPI.com.
// @Tags         weather
// @Accept       json
// @Produce      json
// @Param        city  query     string  true  "City name for weather forecast"
// @Success      200   {object}  Weather  "Successful operation - current weather forecast returned"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "City not found"
// @Router       /weather [get]
func (h *Handler) GetWeather(ctx *gin.Context) {

	city := ctx.Query("city")

	fetchedWeather, err, code := h.srvc.Weather.FetchWeatherForCity(city)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}

	ctx.JSON(200, fetchedWeather)
}
