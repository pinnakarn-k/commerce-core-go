package payment

import "github.com/gin-gonic/gin"

const webhookBasePath = "/webhooks/payments"

func RegisterRoutes(r gin.IRouter, h *Handler) {
	webhooks := r.Group(webhookBasePath)

	webhooks.POST("/mock", h.MockWebhook)
}
