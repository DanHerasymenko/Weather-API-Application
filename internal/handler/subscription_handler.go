package handler

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/services/scheduler_service"
	"Weather-API-Application/internal/services/subscription_service"
	"Weather-API-Application/internal/utils/response"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SubscriptionHandler struct {
	config              *config.Config
	subscriptionService *subscription_service.SubscriptionService
	schedulerService    *scheduler_service.SchedulerService
}

func NewSubscriptionHandler(cfg *config.Config, srvc *subscription_service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		config:  cfg,
		service: srvc,
	}
}

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
// @Description  Subscribes an email_service to weather updates for a specific city with the given frequency.
// @Tags         subscription
// @Accept       application/x-www-form-urlencoded
// @Produce      text/plain
// @Param        email_service formData string true "Email address to subscribe"
// @Param        city formData string true "City for weather updates"
// @Param        frequency formData string true "Frequency of updates (hourly or daily)" Enums(hourly, daily)
// @Success 200 {object} Subscription "Subscription successful. Confirmation email_service sent."
// @Failure      400 {string} string "Invalid input"
// @Failure      409 {string} string "Email already subscribed"
// @Failure      500 {string} string "Internal api error"
// @Router       /subscribe [post]
func (h *SubscriptionHandler) Subscribe(ctx *gin.Context) {
	var req model.Subscription

	if err := h.service.Subscribe(ctx, &req); err != nil {
		switch {
		case errors.Is(err, subscription_service.ErrSubscriptionExists):
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Confirmation email_service sent."})
}

// ConfirmSubscription godoc
// @Summary      Confirm email_service subscription
// @Description  Confirms a subscription using the token sent in the confirmation email_service.
// @Tags         subscription
// @Produce      plain
// @Param        token  path      string  true  "Confirmation token"
// @Success      200    {string}  string  "Subscription confirmed successfully"
// @Failure      400    {string}  string  "Invalid token"
// @Failure      404    {string}  string  "Token not found"
// @Router       /confirm/{token} [get]
func (h *SubscriptionHandler) ConfirmSubscription(ctx *gin.Context) {

	token := ctx.Param("token")

	sub, err := h.subscriptionService.ConfirmSubscription(ctx.Request.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, subscription_service.ErrNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, subscription_service.ErrAlreadyConfirmed):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// запускаємо рутину для кокнретної підтвердженої підписки
	go h.schedulerService.StartRoutine(ctx, sub)

	ctx.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed."})
}

// Unsubscribe godoc
// @Summary      Unsubscribe from weather updates
// @Description  Unsubscribes an email_service from weather updates using the token sent in emails.
// @Tags         subscription
// @Produce      plain
// @Param        token  path      string  true  "Unsubscribe token"
// @Success      200    {string}  string  "Unsubscribed successfully"
// @Failure      400    {string}  string  "Invalid token"
// @Failure      404    {string}  string  "Token not found"
// @Router       /unsubscribe/{token} [get]
func (h *SubscriptionHandler) Unsubscribe(ctx *gin.Context) {

	token := ctx.Param("token")
	code, err := h.srvc.Subscription.Unsubscribe(ctx, token)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}

	ctx.String(200, "Unsubscribed successfully")
}
