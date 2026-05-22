package checkout

import (
	"context"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/cart"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/order"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/payment"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/product"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil"

	"github.com/stretchr/testify/require"
)

func TestService_Checkout_Success_Integration(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	cartRepo, err := cart.NewRepository(db)
	require.NoError(t, err)

	orderRepo, err := order.NewRepository(db)
	require.NoError(t, err)

	productRepo, err := product.NewRepository(db)
	require.NoError(t, err)

	paymentRepo, err := payment.NewRepository(db)
	require.NoError(t, err)

	svc, err := NewService(
		orderRepo,
		cartRepo,
		productRepo,
		paymentRepo,
		db,
	)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db)

	_, err = db.Exec(
		ctx,
		`
		INSERT INTO cart_items (
			user_id,
			product_id,
			quantity,
			is_selected,
			status
		)
		VALUES (
			$1,
			$2,
			2,
			true,
			'active'
		)
		`,
		userID,
		productID,
	)
	require.NoError(t, err)

	result, err := svc.Checkout(ctx, CheckoutCommand{
		UserID:          userID,
		IdempotencyKey:  "idem-001",
		PaymentProvider: "mock",
		PaymentMethod:   "promptpay",
	})

	require.NoError(t, err)

	require.NotZero(t, result.Order.ID)
	require.Equal(t, userID, result.Order.UserID)
	require.Equal(t, "idem-001", result.Order.IdempotencyKey)
	require.Equal(t, 3000, result.Order.TotalAmount)

	require.NotZero(t, result.Payment.ID)
	require.Equal(t, result.Order.ID, result.Payment.OrderID)
	require.Equal(t, payment.PaymentStatusPending, result.Payment.Status)

	var stockQty int
	err = db.QueryRow(
		ctx,
		`
		SELECT stock_qty
		FROM products
		WHERE id = $1
		`,
		productID,
	).Scan(&stockQty)
	require.NoError(t, err)

	require.Equal(t, 8, stockQty)

	var cartStatus string
	var orderID int64

	err = db.QueryRow(
		ctx,
		`
		SELECT status, order_id
		FROM cart_items
		WHERE user_id = $1
		  AND product_id = $2
		`,
		userID,
		productID,
	).Scan(&cartStatus, &orderID)
	require.NoError(t, err)

	require.Equal(t, "purchased", cartStatus)
	require.Equal(t, result.Order.ID, orderID)

	var orderItemCount int
	err = db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM order_items
		WHERE order_id = $1
		`,
		result.Order.ID,
	).Scan(&orderItemCount)
	require.NoError(t, err)

	require.Equal(t, 1, orderItemCount)
}

func TestService_Checkout_IdempotencyHit_Integration(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	cartRepo, err := cart.NewRepository(db)
	require.NoError(t, err)

	orderRepo, err := order.NewRepository(db)
	require.NoError(t, err)

	productRepo, err := product.NewRepository(db)
	require.NoError(t, err)

	paymentRepo, err := payment.NewRepository(db)
	require.NoError(t, err)

	svc, err := NewService(
		orderRepo,
		cartRepo,
		productRepo,
		paymentRepo,
		db,
	)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db)

	_, err = db.Exec(ctx, `
		INSERT INTO cart_items (
			user_id,
			product_id,
			quantity,
			is_selected,
			status
		)
		VALUES (
			$1,
			$2,
			2,
			true,
			'active'
		)
		`,
		userID,
		productID,
	)
	require.NoError(t, err)

	firstResult, err := svc.Checkout(ctx, CheckoutCommand{
		UserID:          userID,
		IdempotencyKey:  "idem-001",
		PaymentProvider: "mock",
		PaymentMethod:   "promptpay",
	})
	require.NoError(t, err)

	secondResult, err := svc.Checkout(ctx, CheckoutCommand{
		UserID:          userID,
		IdempotencyKey:  "idem-001",
		PaymentProvider: "mock",
		PaymentMethod:   "promptpay",
	})
	require.NoError(t, err)

	require.Equal(t, firstResult.Order.ID, secondResult.Order.ID)
	require.Equal(t, firstResult.Payment.ID, secondResult.Payment.ID)

	var stockQty int
	err = db.QueryRow(ctx, `
		SELECT stock_qty
		FROM products
		WHERE id = $1
	`, productID).Scan(&stockQty)
	require.NoError(t, err)

	require.Equal(t, 8, stockQty)

	var orderCount int
	err = db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM orders
		WHERE user_id = $1
	`, userID).Scan(&orderCount)
	require.NoError(t, err)

	require.Equal(t, 1, orderCount)

	var paymentCount int
	err = db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM payments
		WHERE order_id = $1
	`, firstResult.Order.ID).Scan(&paymentCount)
	require.NoError(t, err)

	require.Equal(t, 1, paymentCount)
}

func TestService_Checkout_CartEmpty_Integration(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	cartRepo, err := cart.NewRepository(db)
	require.NoError(t, err)

	orderRepo, err := order.NewRepository(db)
	require.NoError(t, err)

	productRepo, err := product.NewRepository(db)
	require.NoError(t, err)

	paymentRepo, err := payment.NewRepository(db)
	require.NoError(t, err)

	svc, err := NewService(
		orderRepo,
		cartRepo,
		productRepo,
		paymentRepo,
		db,
	)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	result, err := svc.Checkout(ctx, CheckoutCommand{
		UserID:          userID,
		IdempotencyKey:  "idem-empty",
		PaymentProvider: "mock",
		PaymentMethod:   "promptpay",
	})

	require.Nil(t, result)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "CART_EMPTY", appErr.Code)

	var orderCount int
	err = db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM orders
		WHERE user_id = $1
		`,
		userID,
	).Scan(&orderCount)
	require.NoError(t, err)

	require.Equal(t, 0, orderCount)
}

func TestService_Checkout_ProductUnavailable_Rollback(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	cartRepo, err := cart.NewRepository(db)
	require.NoError(t, err)

	orderRepo, err := order.NewRepository(db)
	require.NoError(t, err)

	productRepo, err := product.NewRepository(db)
	require.NoError(t, err)

	paymentRepo, err := payment.NewRepository(db)
	require.NoError(t, err)

	svc, err := NewService(
		orderRepo,
		cartRepo,
		productRepo,
		paymentRepo,
		db,
	)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db, testutil.TestProductOption{
		StockQty: 1,
	})

	_, err = db.Exec(ctx, `
			INSERT INTO cart_items (
				user_id,
				product_id,
				quantity,
				is_selected,
				status
			)
			VALUES (
				$1,
				$2,
				5,
				true,
				'active'
			)
		`,
		userID,
		productID,
	)
	require.NoError(t, err)

	result, err := svc.Checkout(ctx, CheckoutCommand{
		UserID:          userID,
		IdempotencyKey:  "idem-stock-fail",
		PaymentProvider: "mock",
		PaymentMethod:   "promptpay",
	})

	require.Nil(t, result)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "PRODUCT_UNAVAILABLE", appErr.Code)

	var stockQty int
	err = db.QueryRow(ctx, `
		SELECT stock_qty
		FROM products
		WHERE id = $1
	`, productID).Scan(&stockQty)
	require.NoError(t, err)

	require.Equal(t, 1, stockQty)

	var orderCount int
	err = db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM orders
		WHERE user_id = $1
	`, userID).Scan(&orderCount)
	require.NoError(t, err)

	require.Equal(t, 0, orderCount)

	var paymentCount int
	err = db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM payments
	`).Scan(&paymentCount)
	require.NoError(t, err)

	require.Equal(t, 0, paymentCount)

	var cartStatus string
	err = db.QueryRow(ctx, `
		SELECT status
		FROM cart_items
		WHERE user_id = $1
		  AND product_id = $2
	`,
		userID,
		productID,
	).Scan(&cartStatus)
	require.NoError(t, err)

	require.Equal(t, "active", cartStatus)
}
