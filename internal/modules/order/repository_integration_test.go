package order

import (
	"context"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/testutil"

	"github.com/stretchr/testify/require"
)

func TestRepository_CreateOrder_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	o := &Order{
		UserID:         userID,
		IdempotencyKey: "idem-001",
		Status:         OrderStatusPending,
		TotalAmount:    3000,
		Currency:       "THB",
	}

	err = repo.CreateOrder(ctx, tx, o)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	require.NotZero(t, o.ID)
	require.Equal(t, userID, o.UserID)
	require.Equal(t, "idem-001", o.IdempotencyKey)
	require.Equal(t, OrderStatusPending, o.Status)
	require.Equal(t, 3000, o.TotalAmount)
	require.Equal(t, "THB", o.Currency)
	require.Nil(t, o.PaidAt)
	require.Nil(t, o.CancelledAt)
	require.False(t, o.CreatedAt.IsZero())
	require.False(t, o.UpdatedAt.IsZero())
}

func TestRepository_CreateOrderItem_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	productID := testutil.CreateTestProduct(t, db)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	err = repo.CreateOrderItem(ctx, tx, orderID, CreateOrderItemInput{
		ProductID:       productID,
		ProductName:     "Keyboard",
		ProductSKU:      "SKU-001",
		Quantity:        2,
		UnitPriceAmount: 1500,
		LineTotalAmount: 3000,
		Currency:        "THB",
	})
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	var count int
	err = db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM order_items
		WHERE order_id = $1
	`, orderID).Scan(&count)
	require.NoError(t, err)

	require.Equal(t, 1, count)
}

func TestRepository_FindOrderByIdempotencyKey_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID, testutil.TestOrderOption{
		IdempotencyKey: "idem-001",
		TotalAmount:    3000,
	})

	got, err := repo.FindOrderByIdempotencyKey(
		ctx,
		userID,
		"idem-001",
	)

	require.NoError(t, err)
	require.Equal(t, orderID, got.ID)
	require.Equal(t, userID, got.UserID)
	require.Equal(t, "idem-001", got.IdempotencyKey)
	require.Equal(t, OrderStatusPending, got.Status)
	require.Equal(t, 3000, got.TotalAmount)
	require.Equal(t, "THB", got.Currency)
}

func TestRepository_FindOrderByIdempotencyKey_NotFound(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	got, err := repo.FindOrderByIdempotencyKey(
		ctx,
		int64(999),
		"idem-not-found",
	)

	require.Nil(t, got)
	require.ErrorIs(t, err, ErrOrderNotFound)
}

func TestRepository_ListByUserID_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	var otherUserID int64
	err = db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ('Other User', 'other@example.com', 'hashed-password')
		RETURNING id
	`).Scan(&otherUserID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
		INSERT INTO orders (
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency,
			paid_at
		)
		VALUES
			($1, 'idem-001', 'pending', 1000, 'THB', null),
			($1, 'idem-002', 'paid', 2000, 'THB', now()),
			($2, 'idem-003', 'pending', 3000, 'THB', null)
	`,
		userID,
		otherUserID,
	)
	require.NoError(t, err)

	orders, err := repo.ListByUserID(ctx, userID)

	require.NoError(t, err)
	require.Len(t, orders, 2)

	require.Equal(t, userID, orders[0].UserID)
	require.Equal(t, userID, orders[1].UserID)

	require.Equal(t, "idem-001", orders[0].IdempotencyKey)
	require.Equal(t, "idem-002", orders[1].IdempotencyKey)
}

func TestRepository_GetDetailByID_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID, testutil.TestOrderOption{
		IdempotencyKey: "idem-001",
		TotalAmount:    3000,
	})

	keyboardID := testutil.CreateTestProduct(t, db, testutil.TestProductOption{
		SKU:         "SKU-001",
		Name:        "Keyboard",
		PriceAmount: 1500,
		Currency:    "THB",
		StockQty:    10,
	})
	require.NoError(t, err)

	mouseID := testutil.CreateTestProduct(t, db, testutil.TestProductOption{
		SKU:         "SKU-002",
		Name:        "Mouse",
		PriceAmount: 500,
		Currency:    "THB",
		StockQty:    10,
	})
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
		INSERT INTO order_items (
			order_id,
			product_id,
			product_name,
			product_sku,
			quantity,
			unit_price_amount,
			line_total_amount,
			currency
		)
		VALUES
			($1, $2, 'Keyboard', 'SKU-001', 1, 1500, 1500, 'THB'),
			($1, $3, 'Mouse', 'SKU-002', 3, 500, 1500, 'THB')
	`, orderID, keyboardID, mouseID)
	require.NoError(t, err)

	detail, err := repo.GetDetailByID(ctx, userID, orderID)

	require.NoError(t, err)

	require.Equal(t, orderID, detail.Order.ID)
	require.Equal(t, userID, detail.Order.UserID)
	require.Equal(t, "idem-001", detail.Order.IdempotencyKey)

	require.Len(t, detail.Items, 2)

	require.Equal(t, "Keyboard", detail.Items[0].ProductName)
	require.Equal(t, "Mouse", detail.Items[1].ProductName)
}

func TestRepository_GetDetailByID_NotFound(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	detail, err := repo.GetDetailByID(
		ctx,
		int64(1),
		int64(999),
	)

	require.Nil(t, detail)
	require.ErrorIs(t, err, ErrOrderNotFound)
}
