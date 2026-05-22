package auth

import "github.com/gin-gonic/gin"

const basePath = "/auth"

func RegisterRoutes(r gin.IRouter, h *Handler) {
	auth := r.Group(basePath)

	auth.POST("/login", h.Login)
}
