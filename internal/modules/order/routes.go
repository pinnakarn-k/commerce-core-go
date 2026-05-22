package order

import "github.com/gin-gonic/gin"

const basePath = "/me/orders"

func RegisterRoutes(
	r gin.IRouter,
	h *Handler,
	authMiddleware gin.HandlerFunc,
) {
	orders := r.Group(basePath)
	orders.Use(authMiddleware)

	orders.GET("", h.List)
	orders.GET("/:id", h.GetDetailByID)
}
