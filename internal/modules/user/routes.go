package user

import "github.com/gin-gonic/gin"

const basePath = "/users"

func RegisterRoutes(
	r gin.IRouter,
	h *Handler,
	authMiddleware gin.HandlerFunc,
) {
	users := r.Group(basePath)

	users.POST("", h.Create)

	me := r.Group("/me")
	me.Use(authMiddleware)
	me.GET("", h.Me)
}
