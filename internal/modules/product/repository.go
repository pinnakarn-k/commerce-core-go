package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	DeductStock(ctx context.Context, tx pgx.Tx, productID int64, quantity int) error
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

func (r *postgresRepository) DeductStock(ctx context.Context, tx pgx.Tx, productID int64, quantity int) error {
	const query = `
		UPDATE products
		SET stock_qty = stock_qty - $1
		WHERE id = $2
		  AND status = 'active'
		  AND stock_qty >= $1
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(ctx, query, quantity, productID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductUnavailable
		}

		return fmt.Errorf("deduct stock: %w", err)
	}

	return nil
}
