package cart

import (
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
)

var (
	ErrNilDB              = errors.New("cart repository: db is nil")
	ErrNilService         = errors.New("cart handler: service is nil")
	ErrNilRepository      = errors.New("cart service: repository is nil")
	ErrCartEmpty          = errors.New("cart empty")
	ErrCartItemNotFound   = errors.New("cart item not found")
	ErrProductUnavailable = errors.New("product unavailable or stock not enough")
)

func ProductUnavailable(err error) *apperror.AppError {
	return apperror.BadRequest(
		"PRODUCT_UNAVAILABLE",
		"product is unavailable or stock is not enough",
		err,
	)
}

func CartItemNotFound(err error) *apperror.AppError {
	return apperror.NotFound(
		"CART_ITEM_NOT_FOUND",
		"cart item not found",
		err,
	)
}
