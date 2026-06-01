package order

import (
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
)

var (
	ErrNilService               = errors.New("order handler: service is nil")
	ErrNilRepository            = errors.New("nil order repository")
	ErrNilDB                    = errors.New("order repository: db is nil")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderIdempotencyConflict = errors.New("order idempotency conflict")
)

func OrderNotFound(err error) *apperror.AppError {
	return apperror.NotFound(
		"ORDER_NOT_FOUND",
		"order not found",
		err,
	)
}

func CartEmpty(err error) *apperror.AppError {
	return apperror.BadRequest(
		"CART_EMPTY",
		"cart empty",
		err,
	)
}

func ProductUnavailable(err error) *apperror.AppError {
	return apperror.Conflict(
		"PRODUCT_UNAVAILABLE",
		"product unavailable",
		err,
	)
}

func InvalidUserID(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_USER_ID",
		"invalid user id",
		err,
	)
}

func InvalidOrderID(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_ORDER_ID",
		"invalid order id",
		err,
	)
}
