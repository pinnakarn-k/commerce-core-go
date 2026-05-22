package payment

import (
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
)

var (
	ErrNilService                   = errors.New("payment handler: service is nil")
	ErrNilRepository                = errors.New("nil payment repository")
	ErrNilOrderRepo                 = errors.New("nil order repository")
	ErrNilDB                        = errors.New("payment repository: db is nil")
	ErrPaymentNotFound              = errors.New("payment not found")
	ErrPaymentEventAlreadyProcessed = errors.New("payment event already processed")
)

func PaymentNotFound(err error) *apperror.AppError {
	return apperror.NotFound(
		"PAYMENT_NOT_FOUND",
		"payment not found",
		err,
	)
}

func InvalidPaymentID(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_PAYMENT_ID",
		"invalid payment id",
		err,
	)
}

func InvalidPaymentStatus(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_PAYMENT_STATUS",
		"invalid payment status",
		err,
	)
}

func InvalidPaymentEventID(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_PAYMENT_EVENT_ID",
		"invalid payment event id",
		err,
	)
}

func InvalidProviderPaymentID(err error) *apperror.AppError {
	return apperror.BadRequest(
		"INVALID_PROVIDER_PAYMENT_ID",
		"invalid provider payment id",
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
