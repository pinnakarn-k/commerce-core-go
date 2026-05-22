package testpayment

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type CreatePaymentInput struct {
	OrderID           int64
	IdempotencyKey    string
	Provider          string
	Method            string
	ProviderPaymentID string
	Status            string
	Amount            int64
	Currency          string
}

func CreatePayment(
	t *testing.T,
	db *pgxpool.Pool,
	input CreatePaymentInput,
) int64 {
	t.Helper()

	if input.IdempotencyKey == "" {
		input.IdempotencyKey = "idem-001"
	}

	if input.Provider == "" {
		input.Provider = "mock"
	}

	if input.Method == "" {
		input.Method = "promptpay"
	}

	if input.ProviderPaymentID == "" {
		input.ProviderPaymentID = "mock_pay_001"
	}

	if input.Status == "" {
		input.Status = "pending"
	}

	if input.Currency == "" {
		input.Currency = "THB"
	}

	var paymentID int64

	err := db.QueryRow(
		t.Context(),
		`
		INSERT INTO payments (
			order_id,
			idempotency_key,
			provider,
			method,
			provider_payment_id,
			status,
			amount,
			currency
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		)
		RETURNING id
		`,
		input.OrderID,
		input.IdempotencyKey,
		input.Provider,
		input.Method,
		input.ProviderPaymentID,
		input.Status,
		input.Amount,
		input.Currency,
	).Scan(&paymentID)

	require.NoError(t, err)

	return paymentID
}
