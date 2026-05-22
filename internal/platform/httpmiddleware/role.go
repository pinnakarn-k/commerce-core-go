package httpmiddleware

import (
	"net/http"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/authcontext"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/response"

	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		authUser, err := authcontext.Get(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
			c.Abort()
			return
		}

		if _, ok := allowed[authUser.Role]; !ok {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}
