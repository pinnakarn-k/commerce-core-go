package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListByUserID(ctx context.Context, userID int64) ([]Order, error)
	GetDetailByID(ctx context.Context, userID int64, orderID int64) (*OrderDetail, error)
	FindOrderByIdempotencyKey(ctx context.Context, userID int64, idempotencyKey string) (*Order, error)
	CreateOrder(ctx context.Context, tx pgx.Tx, order *Order) error
	CreateOrderItem(ctx context.Context, tx pgx.Tx, orderID int64, input CreateOrderItemInput) error
	MarkPaid(ctx context.Context, tx pgx.Tx, orderID int64) error
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

func (r *postgresRepository) ListByUserID(ctx context.Context, userID int64) ([]Order, error) {
	const query = `
		SELECT
			id,
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency,
			created_at,
			updated_at,
			paid_at,
			cancelled_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list orders by user id: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0)

	for rows.Next() {
		var order Order

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.IdempotencyKey,
			&order.Status,
			&order.TotalAmount,
			&order.Currency,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.PaidAt,
			&order.CancelledAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

func (r *postgresRepository) GetDetailByID(ctx context.Context, userID int64, orderID int64) (*OrderDetail, error) {
	const orderQuery = `
		SELECT
			id,
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency,
			created_at,
			updated_at,
			paid_at,
			cancelled_at
		FROM orders
		WHERE id = $1
		  AND user_id = $2
	`

	order := Order{}

	err := r.db.QueryRow(
		ctx,
		orderQuery,
		orderID,
		userID,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.IdempotencyKey,
		&order.Status,
		&order.TotalAmount,
		&order.Currency,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.PaidAt,
		&order.CancelledAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}

		return nil, fmt.Errorf("get order by id: %w", err)
	}

	const itemsQuery = `
		SELECT
			id,
			order_id,
			product_id,
			product_name,
			product_sku,
			quantity,
			unit_price_amount,
			line_total_amount,
			currency,
			created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, order.ID)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()

	items := make([]OrderItem, 0)

	for rows.Next() {
		var item OrderItem

		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.Quantity,
			&item.UnitPriceAmount,
			&item.LineTotalAmount,
			&item.Currency,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}

	return &OrderDetail{
		Order: order,
		Items: items,
	}, nil
}

func (r *postgresRepository) FindOrderByIdempotencyKey(ctx context.Context, userID int64, idempotencyKey string) (*Order, error) {
	const query = `
		SELECT
			id,
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency,
			created_at,
			updated_at,
			paid_at,
			cancelled_at
		FROM orders
		WHERE user_id = $1
		  AND idempotency_key = $2
	`

	order := &Order{}

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		idempotencyKey,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.IdempotencyKey,
		&order.Status,
		&order.TotalAmount,
		&order.Currency,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.PaidAt,
		&order.CancelledAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}

		return nil, fmt.Errorf("find order by idempotency key: %w", err)
	}

	return order, nil
}

func (r *postgresRepository) CreateOrder(ctx context.Context, tx pgx.Tx, order *Order) error {
	const query = `
		INSERT INTO orders (
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency,
			created_at,
			updated_at,
			paid_at,
			cancelled_at
	`

	err := tx.QueryRow(
		ctx,
		query,
		order.UserID,
		order.IdempotencyKey,
		order.Status,
		order.TotalAmount,
		order.Currency,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.IdempotencyKey,
		&order.Status,
		&order.TotalAmount,
		&order.Currency,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.PaidAt,
		&order.CancelledAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrOrderIdempotencyConflict
		}

		return fmt.Errorf("create order: %w", err)
	}

	return nil
}

func (r *postgresRepository) CreateOrderItem(ctx context.Context, tx pgx.Tx, orderID int64, input CreateOrderItemInput) error {
	const query = `
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
	`

	_, err := tx.Exec(
		ctx,
		query,
		orderID,
		input.ProductID,
		input.ProductName,
		input.ProductSKU,
		input.Quantity,
		input.UnitPriceAmount,
		input.LineTotalAmount,
		input.Currency,
	)
	if err != nil {
		return fmt.Errorf("create order item: %w", err)
	}

	return nil
}

func (r *postgresRepository) MarkPaid(ctx context.Context, tx pgx.Tx, orderID int64) error {
	const query = `
		UPDATE orders
		SET
			status = 'paid',
			paid_at = now(),
			updated_at = now()
		WHERE id = $1
			AND status = 'pending'
	`

	commandTag, err := tx.Exec(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("mark order paid: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrOrderNotFound
	}

	return nil
}
