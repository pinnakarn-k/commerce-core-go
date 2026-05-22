package checkout

import (
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
)

var (
	ErrNilService           = errors.New("order handler: service is nil")
	ErrNilRepository        = errors.New("nil order repository")
	ErrNilCartRepository    = errors.New("nil cart repository")
	ErrNilProductRepository = errors.New("nil product repository")
	ErrNilPaymentRepository = errors.New("nil payment repository")
	ErrNilDB                = errors.New("order repository: db is nil")
	ErrOrderNotFound        = errors.New("order not found")
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

func InvalidIdempotencyKey(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_IDEMPOTENCY_KEY",
		"invalid idempotency key",
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

func InvalidPaymentProvider(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_PAYMENT_PROVIDER",
		"invalid payment provider",
		err,
	)
}

func InvalidPaymentMethod(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_PAYMENT_METHOD",
		"invalid payment method",
		err,
	)
}
