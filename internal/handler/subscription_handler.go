package handler

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/model"
	"Weather-API-Application/internal/services"
	"Weather-API-Application/internal/utils/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SubscriptionHandler struct {
	config  *config.Config
	service *services.SubscriptionService
}

func NewSubscriptionHandler(cfg *config.Config, srvc *services.SubscriptionService) *SubscriptionHandler {
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
// @Failure      500 {string} string "Internal api error"
// @Router       /subscribe [post]
func (h *SubscriptionHandler) Subscribe(ctx *gin.Context) {

	// 1. Використовуємо модель для автоматичного парсингу JSON
	var req model.SubscriptionCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// 2. Уся логіка, включно з валідацією, тепер у сервісі.
	// Обробник просто передає дані.
	if err := h.service.Subscribe(ctx, &req); err != nil {
		// Сервіс повинен повертати помилки, які можна перетворити на HTTP статус.
		// Для простоти, ми повертаємо 500, але можна додати логіку для різних кодів помилок.
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Використовуємо стандартну відповідь Gin
	ctx.String(http.StatusOK, "Subscription successful. Confirmation email sent.")
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
func (h *SubscriptionHandler) ConfirmSubscription(ctx *gin.Context) {

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
func (h *SubscriptionHandler) Unsubscribe(ctx *gin.Context) {

	token := ctx.Param("token")
	code, err := h.srvc.Subscription.Unsubs	cribe(ctx, token)
	if err != nil {
		response.AbortWithError(ctx, code, err)
		return
	}

	ctx.String(200, "Unsubscribed successfully")
}

