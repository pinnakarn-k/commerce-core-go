package user

import (
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
)

var (
	ErrNilDB              = errors.New("user repository: db is nil")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

func InvalidUserID(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_USER_ID",
		"invalid user id",
		err,
	)
}

func UserNotFound(err error) *apperror.AppError {
	return apperror.NotFound(
		"USER_NOT_FOUND",
		"user not found",
		err,
	)
}

func EmailAlreadyExists(err error) *apperror.AppError {
	return apperror.Conflict(
		"EMAIL_ALREADY_EXISTS",
		"email already exists",
		err,
	)
}

func InvalidRole(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_ROLE",
		"invalid role",
		err,
	)
}
