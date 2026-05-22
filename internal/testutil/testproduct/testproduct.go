package testproduct

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type CreateProductInput struct {
	Name        string
	SKU         string
	PriceAmount int64
	StockQty    int
}

func CreateActiveProduct(
	t *testing.T,
	db *pgxpool.Pool,
	input CreateProductInput,
) int64 {
	t.Helper()

	var productID int64

	err := db.QueryRow(
		t.Context(),
		`
		INSERT INTO products (
			name,
			sku,
			price_amount,
			currency,
			stock_qty,
			status
		)
		VALUES (
			$1,
			$2,
			$3,
			'THB',
			$4,
			'active'
		)
		RETURNING id
		`,
		input.Name,
		input.SKU,
		input.PriceAmount,
		input.StockQty,
	).Scan(&productID)

	require.NoError(t, err)

	return productID
}
