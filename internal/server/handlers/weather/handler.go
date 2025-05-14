package weather

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/services"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	cfg      *config.Config
	services *services.Services
}

func NewHandler(cfg *config.Config, services *services.Services) *Handler {
	return &Handler{
		cfg:      cfg,
		services: services,
	}
}

type Weather struct {
	Temperature int    `json:"temperature" example:"25"`    // Current temperature
	Humidity    int    `json:"humidity" example:"60"`       // Current humidity percentage
	Description string `json:"description" example:"Sunny"` // Weather description
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
// @Router       /api/weather [get]
func (h *Handler) GetWeather(ctx *gin.Context) {
	city := ctx.Query("city")

	//

	ctx.String(http.StatusBadRequest, "Invalid request")

	fmt.Println("City:", city)
	ctx.JSON(200, Weather{
		Temperature: 25,
		Humidity:    60,
		Description: "Sunny",
	})
}
