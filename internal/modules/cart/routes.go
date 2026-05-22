package cart

import "github.com/gin-gonic/gin"

const basePath = "/me/cart"

func RegisterRoutes(
	r gin.IRouter,
	h *Handler,
	authMiddleware gin.HandlerFunc,
) {
	cart := r.Group(basePath)

	cart.Use(authMiddleware)

	cart.GET("", h.ListItems)
	cart.PUT("/items", h.UpsertItem)
	cart.DELETE("/items/:product_id", h.RemoveItem)
}
