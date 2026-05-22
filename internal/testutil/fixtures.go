package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func CreateTestUser(t *testing.T, db *pgxpool.Pool) int64 {
	t.Helper()

	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())

	var userID int64
	err := db.QueryRow(
		context.Background(),
		`
		INSERT INTO users (
			name,
			email,
			password_hash
		)
		VALUES (
			'Test User',
			$1,
			'hashed-password'
		)
		RETURNING id
		`,
		email,
	).Scan(&userID)
	require.NoError(t, err)

	return userID
}

type TestProductOption struct {
	SKU         string
	Name        string
	PriceAmount int
	Currency    string
	StockQty    int
	Status      string
}

func CreateTestProduct(
	t *testing.T,
	db *pgxpool.Pool,
	opts ...TestProductOption,
) int64 {
	t.Helper()

	opt := TestProductOption{
		SKU:         fmt.Sprintf("SKU-%d", time.Now().UnixNano()),
		Name:        "Keyboard",
		PriceAmount: 1500,
		Currency:    "THB",
		StockQty:    10,
		Status:      "active",
	}

	if len(opts) > 0 {
		if opts[0].SKU != "" {
			opt.SKU = opts[0].SKU
		}
		if opts[0].Name != "" {
			opt.Name = opts[0].Name
		}
		if opts[0].PriceAmount != 0 {
			opt.PriceAmount = opts[0].PriceAmount
		}
		if opts[0].Currency != "" {
			opt.Currency = opts[0].Currency
		}
		if opts[0].StockQty != 0 {
			opt.StockQty = opts[0].StockQty
		}
		if opts[0].Status != "" {
			opt.Status = opts[0].Status
		}
	}

	var productID int64
	err := db.QueryRow(
		context.Background(),
		`
		INSERT INTO products (
			sku,
			name,
			price_amount,
			currency,
			stock_qty,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`,
		opt.SKU,
		opt.Name,
		opt.PriceAmount,
		opt.Currency,
		opt.StockQty,
		opt.Status,
	).Scan(&productID)
	require.NoError(t, err)

	return productID
}

type TestOrderOption struct {
	IdempotencyKey string
	Status         string
	TotalAmount    int
	Currency       string
}

func CreateTestOrder(
	t *testing.T,
	db *pgxpool.Pool,
	userID int64,
	opts ...TestOrderOption,
) int64 {
	t.Helper()

	opt := TestOrderOption{
		IdempotencyKey: fmt.Sprintf("idem-%d", time.Now().UnixNano()),
		Status:         "pending",
		TotalAmount:    3000,
		Currency:       "THB",
	}

	if len(opts) > 0 {
		if opts[0].IdempotencyKey != "" {
			opt.IdempotencyKey = opts[0].IdempotencyKey
		}
		if opts[0].Status != "" {
			opt.Status = opts[0].Status
		}
		if opts[0].TotalAmount != 0 {
			opt.TotalAmount = opts[0].TotalAmount
		}
		if opts[0].Currency != "" {
			opt.Currency = opts[0].Currency
		}
	}

	var orderID int64
	err := db.QueryRow(
		context.Background(),
		`
		INSERT INTO orders (
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
		`,
		userID,
		opt.IdempotencyKey,
		opt.Status,
		opt.TotalAmount,
		opt.Currency,
	).Scan(&orderID)
	require.NoError(t, err)

	return orderID
}
