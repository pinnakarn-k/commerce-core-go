package authcontext

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const AuthUserKey = "auth_user"

var ErrAuthUserNotFound = errors.New("auth user not found")

type AuthUser struct {
	ID   int64
	Role string
}

func Set(c *gin.Context, authUser AuthUser) {
	c.Set(AuthUserKey, authUser)
}

func Get(c *gin.Context) (AuthUser, error) {
	v, ok := c.Get(AuthUserKey)
	if !ok {
		return AuthUser{}, ErrAuthUserNotFound
	}

	authUser, ok := v.(AuthUser)
	if !ok {
		return AuthUser{}, ErrAuthUserNotFound
	}

	return authUser, nil
}
