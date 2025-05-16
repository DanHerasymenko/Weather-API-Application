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

type SubscribeReqBody struct {
	email      string `json:"email" binding:"required,email"`
	city       string `json:"city" binding:"required,alphanum"`
	fruequency string `json:"frequency" binding:"required,oneof=hourly daily"`
}

// Subscribe godoc
func (h *Handler) Subscribe(ctx *gin.Context) {
	reqBody := SubscribeReqBody{}

	if err := validate.ParseReqBody(ctx, &reqBody); err != nil {
		response.AbortWithErrorJSON(ctx, http.StatusBadRequest, err, "Email or password is invalid")
		return
	}

	ctx.String(200, "Subscription successful. Confirmation email sent.")
}
