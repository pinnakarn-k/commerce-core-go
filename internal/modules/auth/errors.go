package auth

import (
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
)

var (
	ErrNilService     = errors.New("auth handler: service is nil")
	ErrUserRepository = errors.New("auth service: user repository is nil")
	ErrTokenMaker     = errors.New("auth service: token maker is nil")
)

func InvalidCredentials(err error) *apperror.AppError {
	return apperror.Unauthorized(
		"INVALID_CREDENTIALS",
		"invalid email or password",
		err,
	)
}

func UserInactive(err error) *apperror.AppError {
	return apperror.Forbidden(
		"USER_INACTIVE",
		"user is inactive",
		err,
	)
}
