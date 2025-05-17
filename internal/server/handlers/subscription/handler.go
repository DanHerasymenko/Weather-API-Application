package subscription

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/services"
	"Weather-API-Application/internal/utils/response"
	"Weather-API-Application/internal/utils/validate"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
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

// ConfirmSubscription godoc
// @Summary      Confirm email subscription
// @Description  Confirms a subscription using the token sent in the confirmation email.
// @Tags         subscription
// @Produce      plain
// @Param        token  path      string  true  "Confirmation token"
// @Success      200    {string}  string  "Subscription confirmed successfully"
// @Failure      400    {string}  string  "Invalid token"
// @Failure      404    {string}  string  "Token not found"
// @Router       /api/subscription/confirm/{token} [get]
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
// @Router       /api/subscription/unsubscribe/{token} [get]
func (h *Handler) Unsubscribe(ctx *gin.Context) {

	token := ctx.Param("token")
	code, err := h.srvc.Subscription.Unsubscribe(ctx, token)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}

	ctx.String(200, "Subscription confirmed")
}

// SubscribeReqBody represents the expected payload for a subscription request.
type SubscribeReqBody struct {
	Email     string `json:"email" validate:"required,email"`
	City      string `json:"city" validate:"required,alphanum"`
	Frequency string `json:"frequency" validate:"required,oneof=hourly daily"`
}

// Subscribe godoc
// @Summary      Subscribe to weather updates
// @Description  Subscribes an email to weather updates for a specific city with the given frequency.
// @Tags         subscription
// @Accept       json
// @Produce      plain
// @Param        email      body  string  true  "Email address to subscribe"
// @Param        city       body  string  true  "City for weather updates"
// @Param        frequency  body  string  true  "Update frequency (daily or hourly)"
// @Success      200        {string}  string  "Subscription successful. Confirmation email sent."
// @Failure      400        {string}  string  "Invalid input"
// @Failure      409        {string}  string  "Subscription already exists"
// @Failure      500        {string}  string  "Internal server error"
// @Router       /api/subscribe [post]
func (h *Handler) Subscribe(ctx *gin.Context) {
	reqBody := SubscribeReqBody{}

	if err := validate.ParseReqBody(ctx, &reqBody); err != nil {
		response.AbortWithErrorJSON(ctx, http.StatusBadRequest, err, "Email or password is invalid")
		return
	}

	code, err := h.srvc.Subscription.Subscribe(ctx, reqBody.Email, reqBody.City, reqBody.Frequency)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}

	ctx.String(200, "Subscription successful. Confirmation email sent.")
}
