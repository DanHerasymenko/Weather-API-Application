package subscription

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/services"
	"Weather-API-Application/internal/utils/response"
	"Weather-API-Application/internal/utils/validate"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg  *config.Config
	srvc *services.Services
}

func NewHandler(cfg *config.Config, srvc *services.Services) *Handler {
	return &Handler{
		cfg:  cfg,
		srvc: srvc,
	}
}

// swagger:model Subscription
type Subscription struct {
	// Email address
	Email string `json:"email"`

	// City for weather updates
	City string `json:"city"`

	// Frequency of updates
	// Enum: hourly, daily
	Frequency string `json:"frequency"`

	// Whether the subscription is confirmed
	Confirmed bool `json:"confirmed"`
}

// Subscribe godoc
// @Summary      Subscribe to weather updates
// @Description  Subscribes an email to weather updates for a specific city with the given frequency.
// @Tags         subscription
// @Accept       application/x-www-form-urlencoded
// @Produce      text/plain
// @Param        email formData string true "Email address to subscribe"
// @Param        city formData string true "City for weather updates"
// @Param        frequency formData string true "Frequency of updates (hourly or daily)" Enums(hourly, daily)
// @Success 200 {object} Subscription "Subscription successful. Confirmation email sent."
// @Failure      400 {string} string "Invalid input"
// @Failure      409 {string} string "Email already subscribed"
// @Failure      500 {string} string "Internal server error"
// @Router       /subscribe [post]
func (h *Handler) Subscribe(ctx *gin.Context) {

	email := ctx.PostForm("email")
	city := ctx.PostForm("city")
	frequency := ctx.PostForm("frequency")

	if email == "" || city == "" {
		response.AbortWithErrorJSON(ctx, 400, fmt.Errorf("missing fields"), "Email and city are required")
		return
	}
	if frequency != "hourly" && frequency != "daily" {
		response.AbortWithErrorJSON(ctx, 400, fmt.Errorf("invalid frequency"), "Frequency must be 'hourly' or 'daily'")
		return
	}
	if !validate.IsValidEmail(email) {
		response.AbortWithErrorJSON(ctx, 400, fmt.Errorf("invalid email"), "Email format is invalid")
		return
	}

	reqBody := Subscription{
		Email:     email,
		City:      city,
		Frequency: frequency,
	}
	logger.Info(ctx, fmt.Sprintf("VALIDATE INPUT: email=%s | city=%s | frequency=%s", email, city, frequency))

	// check if the city input from User is valid via WeatherAPI
	ok, err, code := h.srvc.Subscription.ValidateCity(reqBody.City)
	logger.Info(ctx, fmt.Sprintf("VALIDATE CITY: %s | ok=%v | err=%v | code=%d", city, ok, err, code))

	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}
	if !ok {
		response.AbortWithError(ctx, code, fmt.Errorf("city not found"))
		return
	}

	code, err = h.srvc.Subscription.Subscribe(ctx, reqBody.Email, reqBody.City, reqBody.Frequency)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}

	ctx.String(200, "Subscription successful. Confirmation email sent.")
}

// ConfirmSubscription godoc
// @Summary      Confirm email subscription
// @Description  Confirms a subscription using the token sent in the confirmation email.
// @Tags         subscription
// @Produce      plain
// @Param        token  path      string  true  "Confirmation token"
// @Success      200    {string}  string  "Subscription confirmed successfully"
// @Failure      400    {string}  string  "Invalid token"
// @Failure      404    {string}  string  "Token not found"
// @Router       /confirm/{token} [get]
func (h *Handler) ConfirmSubscription(ctx *gin.Context) {

	token := ctx.Param("token")
	if token == "" {
		response.AbortWithError(ctx, 400, fmt.Errorf("token is required"))
		return
	}

	code, err := h.srvc.Subscription.ConfirmSubscription(ctx, token)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}

	ctx.String(200, "Subscription confirmed")
}

// Unsubscribe godoc
// @Summary      Unsubscribe from weather updates
// @Description  Unsubscribes an email from weather updates using the token sent in emails.
// @Tags         subscription
// @Produce      plain
// @Param        token  path      string  true  "Unsubscribe token"
// @Success      200    {string}  string  "Unsubscribed successfully"
// @Failure      400    {string}  string  "Invalid token"
// @Failure      404    {string}  string  "Token not found"
// @Router       /unsubscribe/{token} [get]
func (h *Handler) Unsubscribe(ctx *gin.Context) {

	token := ctx.Param("token")
	code, err := h.srvc.Subscription.Unsubscribe(ctx, token)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}

	ctx.String(200, "Unsubscribed successfully")
}
