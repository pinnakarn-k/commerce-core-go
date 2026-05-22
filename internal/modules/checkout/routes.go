package checkout

import "github.com/gin-gonic/gin"

const basePath = "/me/orders"

func RegisterRoutes(
	r gin.IRouter,
	h *Handler,
	authMiddleware gin.HandlerFunc,
) {
	checkout := r.Group(basePath)
	checkout.Use(authMiddleware)

	checkout.POST("/checkout", h.Checkout)
}
