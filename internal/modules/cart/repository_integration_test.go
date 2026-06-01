package cart

import (
	"context"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/testutil"

	"github.com/stretchr/testify/require"
)

func TestRepository_UpsertItem_Success(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db)

	item := &CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  2,
	}

	err = repo.UpsertItem(ctx, item)

	require.NoError(t, err)
	require.NotZero(t, item.ID)
	require.Equal(t, userID, item.UserID)
	require.Equal(t, productID, item.ProductID)
	require.Equal(t, 2, item.Quantity)
	require.True(t, item.IsSelected)
	require.Equal(t, "active", item.Status)
	require.Nil(t, item.OrderID)
	require.False(t, item.CreatedAt.IsZero())
	require.False(t, item.UpdatedAt.IsZero())
}

func TestRepository_UpsertItem_ProductUnavailable_WhenProductNotFound(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	item := &CartItem{
		UserID:    userID,
		ProductID: 999,
		Quantity:  1,
	}

	err = repo.UpsertItem(ctx, item)

	require.ErrorIs(t, err, ErrProductUnavailable)
}

func TestRepository_UpsertItem_ProductUnavailable_WhenStockNotEnough(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db, testutil.TestProductOption{
		StockQty: 1,
	})

	item := &CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  5,
	}

	err = repo.UpsertItem(ctx, item)

	require.ErrorIs(t, err, ErrProductUnavailable)
}

func TestRepository_UpsertItem_ProductUnavailable_WhenProductDisabled(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db, testutil.TestProductOption{
		Status: "disabled",
	})

	item := &CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  1,
	}

	err = repo.UpsertItem(ctx, item)

	require.ErrorIs(t, err, ErrProductUnavailable)
}

func TestRepository_UpsertItem_UpdateExistingActiveItem(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db)

	firstItem := &CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  1,
	}

	err = repo.UpsertItem(ctx, firstItem)
	require.NoError(t, err)

	secondItem := &CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  4,
	}

	err = repo.UpsertItem(ctx, secondItem)
	require.NoError(t, err)

	require.Equal(t, firstItem.ID, secondItem.ID)
	require.Equal(t, 4, secondItem.Quantity)

	var count int
	err = db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM cart_items
		WHERE user_id = $1
			AND product_id = $2
			AND status = 'active'
		`,
		userID,
		productID,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestRepository_ListItems_ReturnsOnlyActiveItemsForUser(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	var otherUserID int64
	err = db.QueryRow(
		ctx,
		`
		INSERT INTO users (name, email, password_hash)
		VALUES ('Other User', 'other@example.com', 'hashed-password')
		RETURNING id
		`,
	).Scan(&otherUserID)
	require.NoError(t, err)

	productID := testutil.CreateTestProduct(t, db)

	otherProductID := testutil.CreateTestProduct(t, db, testutil.TestProductOption{
		SKU:         "SKU-002",
		Name:        "Mouse",
		PriceAmount: 500,
		Currency:    "THB",
		StockQty:    10,
	})

	require.NoError(t, err)

	orderID := testutil.CreateTestOrder(t, db, userID)

	_, err = db.Exec(
		ctx,
		`
		INSERT INTO cart_items (user_id, product_id, quantity, status, order_id)
		VALUES
			($1, $2, 2, 'active', null),
			($1, $3, 1, 'purchased', $4),
			($5, $2, 3, 'active', null)
		`,
		userID,
		productID,
		otherProductID,
		orderID,
		otherUserID,
	)
	require.NoError(t, err)

	items, err := repo.ListItems(ctx, userID)

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, userID, items[0].UserID)
	require.Equal(t, productID, items[0].ProductID)
	require.Equal(t, 2, items[0].Quantity)
	require.Equal(t, "active", items[0].Status)
}

func TestRepository_ListCheckoutItems_ReturnsOnlySelectedActiveItems(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db, testutil.TestProductOption{
		SKU:  "SKU-001",
		Name: "Keyboard",
	})

	unselectedProductID := testutil.CreateTestProduct(t, db, testutil.TestProductOption{
		SKU:         "SKU-002",
		Name:        "Mouse",
		PriceAmount: 500,
		Currency:    "THB",
		StockQty:    10,
	})

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
		VALUES
			($1, $2, 2, true, 'active'),
			($1, $3, 1, false, 'active')
		`,
		userID,
		productID,
		unselectedProductID,
	)
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	items, err := repo.ListCheckoutItems(ctx, tx, userID)
	require.NoError(t, err)

	require.Len(t, items, 1)

	require.Equal(t, productID, items[0].ProductID)
	require.Equal(t, 2, items[0].Quantity)
	require.Equal(t, "Keyboard", items[0].ProductName)
	require.Equal(t, "SKU-001", items[0].ProductSKU)
	require.Equal(t, 1500, items[0].UnitPriceAmount)
	require.Equal(t, "THB", items[0].Currency)
	require.Equal(t, 3000, items[0].LineTotalAmount)
}

func TestRepository_RemoveItem_Success(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db)

	_, err = db.Exec(
		ctx,
		`
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, 2)
		`,
		userID,
		productID,
	)
	require.NoError(t, err)

	err = repo.RemoveItem(ctx, userID, productID)

	require.NoError(t, err)

	var count int
	err = db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM cart_items
		WHERE user_id = $1
			AND product_id = $2
			AND status = 'active'
		`,
		userID,
		productID,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestRepository_RemoveItem_NotFound(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	err = repo.RemoveItem(ctx, int64(1), int64(999))

	require.ErrorIs(t, err, ErrCartItemNotFound)
}

func TestRepository_MarkItemsPurchased_Success(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	productID := testutil.CreateTestProduct(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	var cartItemID int64

	err = db.QueryRow(
		ctx,
		`
		INSERT INTO cart_items (
			user_id,
			product_id,
			quantity,
			is_selected
		)
		VALUES ($1, $2, 2, true)
		RETURNING id
		`,
		userID,
		productID,
	).Scan(&cartItemID)
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	err = repo.MarkItemsPurchased(ctx, tx, userID, orderID, []int64{cartItemID})
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	var status string
	var gotOrderID int64
	err = db.QueryRow(
		ctx,
		`
		SELECT status, order_id
		FROM cart_items
		WHERE id = $1
		`,
		cartItemID,
	).Scan(&status, &gotOrderID)
	require.NoError(t, err)

	require.Equal(t, "purchased", status)
	require.Equal(t, orderID, gotOrderID)
}

func TestRepository_MarkItemsPurchased_CartEmpty(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = repo.MarkItemsPurchased(ctx, tx, userID, orderID, []int64{})

	require.ErrorIs(t, err, ErrCartEmpty)
}
