package handler

import (
	"errors"
	"fmt"
	"net/http"

	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/services/subscription_service"
	"Weather-API-Application/internal/utils/response"
	"Weather-API-Application/internal/utils/validate"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	config              *config.Config
	subscriptionService *subscription_service.SubscriptionService
}

func NewSubscriptionHandler(cfg *config.Config, subSvc *subscription_service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		config:              cfg,
		subscriptionService: subSvc,
	}
}

// RegisterRoutes registers subscription endpoints.
func (h *SubscriptionHandler) RegisterRoutes(router *gin.Engine) {
	subscription := router.Group("/api/subscription")
	{
		subscription.POST("/subscribe", h.Subscribe)
		subscription.GET("/confirm/:token", h.ConfirmSubscription)
		subscription.GET("/unsubscribe/:token", h.Unsubscribe)
	}
}

// Subscribe godoc
// @Summary      Subscribe to weather updates
// @Description  Subscribes an email to weather updates for a city with a frequency.
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        subscription  body   model.Subscription  true  "Subscription request"
// @Success      200  {object}  model.Subscription  "Subscription request accepted. Confirmation email sent."
// @Failure      400  {object}  response.ErrorResponse  "Invalid input"
// @Failure      409  {object}  response.ErrorResponse  "Email already subscribed"
// @Failure      500  {object}  response.ErrorResponse  "Internal error"
// @Router       /subscription/subscribe [post]
func (h *SubscriptionHandler) Subscribe(ctx *gin.Context) {
	var req model.Subscription
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteErrorJSON(ctx, http.StatusBadRequest, err, "Invalid input")
		return
	}

	// Validate input
	if !validate.IsValidEmail(req.Email) {
		response.WriteErrorJSON(ctx, http.StatusBadRequest,
			fmt.Errorf("invalid email format"),
			"Invalid email format")
		return
	}
	if !validate.IsValidCity(req.City) {
		response.WriteErrorJSON(ctx, http.StatusBadRequest,
			fmt.Errorf("invalid city"),
			"City is required and cannot be empty")
		return
	}
	if !validate.IsValidFrequency(req.Frequency) {
		response.WriteErrorJSON(ctx, http.StatusBadRequest,
			fmt.Errorf("invalid frequency"),
			"Frequency must be 'hourly' or 'daily'")
		return
	}

	if err := h.subscriptionService.Subscribe(ctx.Request.Context(), &req); err != nil {
		switch {
		case errors.Is(err, subscription_service.ErrSubscriptionExists):
			response.WriteErrorJSON(ctx, http.StatusConflict, err, "Email already subscribed")
			return
		default:
			response.WriteErrorJSON(ctx, http.StatusInternalServerError, err, "Internal server error")
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Confirmation email sent."})
}

// ConfirmSubscription godoc
// @Summary      Confirm subscription
// @Description  Confirms a subscription using the token from the confirmation email.
// @Tags         subscription
// @Produce      json
// @Param        token  path      string  true  "Confirmation token"
// @Success      200    {string}  string  "Subscription confirmed successfully"
// @Failure      400    {object}  response.ErrorResponse  "Invalid token"
// @Failure      404    {object}  response.ErrorResponse  "Token not found"
// @Router       /subscription/confirm/{token} [get]
func (h *SubscriptionHandler) ConfirmSubscription(ctx *gin.Context) {
	token := ctx.Param("token")

	_, err := h.subscriptionService.ConfirmSubscription(ctx.Request.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, subscription_service.ErrNotFound):
			response.WriteErrorJSON(ctx, http.StatusNotFound, err, "Token not found")
			return
		case errors.Is(err, subscription_service.ErrAlreadyConfirmed):
			response.WriteErrorJSON(ctx, http.StatusBadRequest, err, "Already confirmed")
			return
		default:
			response.WriteErrorJSON(ctx, http.StatusInternalServerError, err, "Internal server error")
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed."})
}

// Unsubscribe godoc
// @Summary      Unsubscribe from weather updates
// @Description  Unsubscribes an email using the provided token.
// @Tags         subscription
// @Produce      json
// @Param        token  path      string  true  "Unsubscribe token"
// @Success      200    {string}  string  "Unsubscribed successfully"
// @Failure      400    {object}  response.ErrorResponse  "Invalid token"
// @Failure      404    {object}  response.ErrorResponse  "Token not found"
// @Router       /subscription/unsubscribe/{token} [get]
func (h *SubscriptionHandler) Unsubscribe(ctx *gin.Context) {
	token := ctx.Param("token")
	if err := h.subscriptionService.Unsubscribe(ctx.Request.Context(), token); err != nil {
		switch {
		case errors.Is(err, subscription_service.ErrNotFound):
			response.WriteErrorJSON(ctx, http.StatusNotFound, err, "Token not found")
			return
		default:
			response.WriteErrorJSON(ctx, http.StatusInternalServerError, err, "Internal server error")
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}
