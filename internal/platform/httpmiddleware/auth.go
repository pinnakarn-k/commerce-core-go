package httpmiddleware

import (
	"net/http"
	"strings"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/authcontext"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/response"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/token"

	"github.com/gin-gonic/gin"
)

type TokenVerifier interface {
	VerifyToken(tokenString string) (*token.Claims, error)
}

func RequireAuth(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header")
			c.Abort()
			return
		}

		claims, err := verifier.VerifyToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			c.Abort()
			return
		}

		authcontext.Set(c, authcontext.AuthUser{
			ID:   claims.UserID,
			Role: string(claims.Role),
		})

		c.Next()
	}
}
