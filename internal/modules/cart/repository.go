package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListItems(ctx context.Context, userID int64) ([]CartItem, error)
	ListCheckoutItems(ctx context.Context, tx pgx.Tx, userID int64) ([]CheckoutItem, error)
	UpsertItem(ctx context.Context, item *CartItem) error
	RemoveItem(ctx context.Context, userID int64, productID int64) error
	MarkItemsPurchased(ctx context.Context, tx pgx.Tx, userID int64, orderID int64, cartItemIDs []int64) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) (Repository, error) {
	if db == nil {
		return nil, ErrNilDB
	}

	return &postgresRepository{db: db}, nil
}

func (r *postgresRepository) ListItems(ctx context.Context, userID int64) ([]CartItem, error) {
	const query = `
		SELECT
			id,
			user_id,
			product_id,
			quantity,
			is_selected,
			status,
			order_id,
			created_at,
			updated_at
		FROM cart_items
		WHERE user_id = $1
		  	AND status = 'active'
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list cart items: %w", err)
	}
	defer rows.Close()

	items := make([]CartItem, 0)

	for rows.Next() {
		var item CartItem

		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ProductID,
			&item.Quantity,
			&item.IsSelected,
			&item.Status,
			&item.OrderID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan cart item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cart items: %w", err)
	}

	return items, nil
}

func (r *postgresRepository) ListCheckoutItems(ctx context.Context, tx pgx.Tx, userID int64) ([]CheckoutItem, error) {
	const query = `
		SELECT
			ci.id AS cart_item_id,
			ci.product_id,
			ci.quantity,
			p.name AS product_name,
			p.sku AS product_sku,
			p.price_amount AS unit_price_amount,
			p.currency,
			ci.quantity * p.price_amount AS line_total_amount
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.user_id = $1
			AND ci.status = 'active'
			AND ci.is_selected = true
		ORDER BY ci.id ASC
		FOR UPDATE OF ci
	`

	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list checkout items: %w", err)
	}
	defer rows.Close()

	items := make([]CheckoutItem, 0)

	for rows.Next() {
		var item CheckoutItem

		if err := rows.Scan(
			&item.CartItemID,
			&item.ProductID,
			&item.Quantity,
			&item.ProductName,
			&item.ProductSKU,
			&item.UnitPriceAmount,
			&item.Currency,
			&item.LineTotalAmount,
		); err != nil {
			return nil, fmt.Errorf("scan checkout item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate checkout items: %w", err)
	}

	return items, nil
}

func (r *postgresRepository) UpsertItem(ctx context.Context, item *CartItem) error {
	const query = `
		INSERT INTO cart_items (
			user_id,
			product_id,
			quantity
		)
		SELECT
			$1,
			p.id,
			$2
		FROM products p
		WHERE p.id = $3
			AND p.status = 'active'
			AND p.stock_qty >= $2
		ON CONFLICT (user_id, product_id)
		WHERE status = 'active'
		DO UPDATE
		SET quantity = EXCLUDED.quantity
		WHERE EXISTS (
			SELECT 1
			FROM products p
			WHERE p.id = cart_items.product_id
				AND p.status = 'active'
				AND p.stock_qty >= EXCLUDED.quantity
		)
		RETURNING 
			id, 
			user_id, 
			product_id, 
			quantity, 
			is_selected, 
			status, 
			order_id, 
			created_at, 
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		item.UserID,
		item.Quantity,
		item.ProductID,
	).Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.Quantity,
		&item.IsSelected,
		&item.Status,
		&item.OrderID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductUnavailable
		}
		return fmt.Errorf("upsert cart item: %w", err)
	}

	return nil
}

func (r *postgresRepository) RemoveItem(ctx context.Context, userID int64, productID int64) error {
	const query = `
		DELETE FROM cart_items
		WHERE user_id = $1
			AND product_id = $2
		  	AND status = 'active'
	`

	res, err := r.db.Exec(ctx, query, userID, productID)
	if err != nil {
		return fmt.Errorf("remove cart item: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrCartItemNotFound
	}

	return nil
}

func (r *postgresRepository) MarkItemsPurchased(ctx context.Context, tx pgx.Tx, userID int64, orderID int64, cartItemIDs []int64) error {
	if len(cartItemIDs) == 0 {
		return ErrCartEmpty
	}

	const query = `
		UPDATE cart_items
		SET
			status = 'purchased',
			order_id = $1
		WHERE user_id = $2
			AND id = ANY($3)
			AND status = 'active'
	`

	res, err := tx.Exec(ctx, query, orderID, userID, cartItemIDs)
	if err != nil {
		return fmt.Errorf("mark cart items purchased: %w", err)
	}

	if res.RowsAffected() != int64(len(cartItemIDs)) {
		return ErrCartEmpty
	}

	return nil
}
