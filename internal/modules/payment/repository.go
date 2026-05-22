package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreatePayment(ctx context.Context, tx pgx.Tx, payment *Payment) error
	FindByOrderIDAndIdempotencyKey(ctx context.Context, orderID int64, idempotencyKey string) (*Payment, error)
	GetByID(ctx context.Context, id int64) (*Payment, error)

	MarkSucceeded(ctx context.Context, tx pgx.Tx, id int64) error
	MarkOrderPaid(ctx context.Context, tx pgx.Tx, orderID int64) error
	MarkFailed(ctx context.Context, tx pgx.Tx, id int64, reason string) error
	MarkCancelled(ctx context.Context, tx pgx.Tx, id int64) error
	MarkExpired(ctx context.Context, tx pgx.Tx, id int64) error

	CreateEvent(ctx context.Context, tx pgx.Tx, event *PaymentEvent) error
	GetByProviderPaymentID(ctx context.Context, provider string, providerPaymentID string) (*Payment, error)
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

func (r *postgresRepository) CreatePayment(ctx context.Context, tx pgx.Tx, payment *Payment) error {
	const query = `
		INSERT INTO payments (
			order_id,
			idempotency_key,
			provider,
			method,
			provider_payment_id,
			status,
			amount,
			currency,
			payment_url,
			qr_code_url,
			failure_reason,
			paid_at,
			failed_at,
			cancelled_at,
			expired_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15
		)
		RETURNING
			id,
			order_id,
			idempotency_key,
			provider,
			method,
			provider_payment_id,
			status,
			amount,
			currency,
			payment_url,
			qr_code_url,
			failure_reason,
			paid_at,
			failed_at,
			cancelled_at,
			expired_at,
			created_at,
			updated_at
	`

	err := tx.QueryRow(
		ctx,
		query,
		payment.OrderID,
		payment.IdempotencyKey,
		payment.Provider,
		payment.Method,
		payment.ProviderPaymentID,
		payment.Status,
		payment.Amount,
		payment.Currency,
		payment.PaymentURL,
		payment.QRCodeURL,
		payment.FailureReason,
		payment.PaidAt,
		payment.FailedAt,
		payment.CancelledAt,
		payment.ExpiredAt,
	).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.IdempotencyKey,
		&payment.Provider,
		&payment.Method,
		&payment.ProviderPaymentID,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.PaymentURL,
		&payment.QRCodeURL,
		&payment.FailureReason,
		&payment.PaidAt,
		&payment.FailedAt,
		&payment.CancelledAt,
		&payment.ExpiredAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}

	return nil
}

func (r *postgresRepository) FindByOrderIDAndIdempotencyKey(ctx context.Context, orderID int64, idempotencyKey string) (*Payment, error) {
	const query = `
		SELECT
			id,
			order_id,
			idempotency_key,
			provider,
			method,
			provider_payment_id,
			status,
			amount,
			currency,
			payment_url,
			qr_code_url,
			failure_reason,
			paid_at,
			failed_at,
			cancelled_at,
			expired_at,
			created_at,
			updated_at
		FROM payments
		WHERE order_id = $1
			AND idempotency_key = $2
	`

	var payment Payment

	err := r.db.QueryRow(
		ctx,
		query,
		orderID,
		idempotencyKey,
	).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.IdempotencyKey,
		&payment.Provider,
		&payment.Method,
		&payment.ProviderPaymentID,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.PaymentURL,
		&payment.QRCodeURL,
		&payment.FailureReason,
		&payment.PaidAt,
		&payment.FailedAt,
		&payment.CancelledAt,
		&payment.ExpiredAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf(
			"find payment by order id and idempotency key: %w",
			err,
		)
	}

	return &payment, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id int64) (*Payment, error) {
	const query = `
		SELECT
			id,
			order_id,
			idempotency_key,
			provider,
			method,
			provider_payment_id,
			status,
			amount,
			currency,
			payment_url,
			qr_code_url,
			failure_reason,
			paid_at,
			failed_at,
			cancelled_at,
			expired_at,
			created_at,
			updated_at
		FROM payments
		WHERE id = $1
	`

	var payment Payment

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.IdempotencyKey,
		&payment.Provider,
		&payment.Method,
		&payment.ProviderPaymentID,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.PaymentURL,
		&payment.QRCodeURL,
		&payment.FailureReason,
		&payment.PaidAt,
		&payment.FailedAt,
		&payment.CancelledAt,
		&payment.ExpiredAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get payment by id: %w", err)
	}

	return &payment, nil
}

func (r *postgresRepository) MarkSucceeded(ctx context.Context, tx pgx.Tx, id int64) error {
	const query = `
		UPDATE payments
		SET
			status = 'succeeded',
			paid_at = now(),
			updated_at = now()
		WHERE id = $1
			AND status = 'pending'
	`

	commandTag, err := tx.Exec(
		ctx,
		query,
		id,
	)
	if err != nil {
		return fmt.Errorf("mark payment succeeded: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

func (r *postgresRepository) MarkOrderPaid(
	ctx context.Context,
	tx pgx.Tx,
	orderID int64,
) error {
	const query = `
		UPDATE orders
		SET
			status = 'paid',
			paid_at = now(),
			updated_at = now()
		WHERE id = $1
			AND status = 'pending'
	`

	commandTag, err := tx.Exec(
		ctx,
		query,
		orderID,
	)
	if err != nil {
		return fmt.Errorf("mark order paid: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

func (r *postgresRepository) MarkFailed(ctx context.Context, tx pgx.Tx, id int64, reason string) error {
	const query = `
		UPDATE payments
		SET
			status = 'failed',
			failure_reason = $2,
			failed_at = now(),
			updated_at = now()
		WHERE id = $1
			AND status = 'pending'
	`

	tag, err := tx.Exec(ctx, query, id, reason)
	if err != nil {
		return fmt.Errorf("mark payment failed: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

func (r *postgresRepository) MarkCancelled(ctx context.Context, tx pgx.Tx, id int64) error {
	const query = `
		UPDATE payments
		SET
			status = 'cancelled',
			cancelled_at = now(),
			updated_at = now()
		WHERE id = $1
			AND status = 'pending'
	`

	tag, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark payment cancelled: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

func (r *postgresRepository) MarkExpired(ctx context.Context, tx pgx.Tx, id int64) error {
	const query = `
		UPDATE payments
		SET
			status = 'expired',
			expired_at = now(),
			updated_at = now()
		WHERE id = $1
			AND status = 'pending'
	`

	tag, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark payment expired: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

func (r *postgresRepository) CreateEvent(ctx context.Context, tx pgx.Tx, event *PaymentEvent) error {
	const query = `
		INSERT INTO payment_events (
			provider,
			provider_event_id,
			payment_id,
			event_type,
			payload
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)
		RETURNING
			id,
			provider,
			provider_event_id,
			payment_id,
			event_type,
			payload,
			created_at
	`

	err := tx.QueryRow(
		ctx,
		query,
		event.Provider,
		event.ProviderEventID,
		event.PaymentID,
		event.EventType,
		event.Payload,
	).Scan(
		&event.ID,
		&event.Provider,
		&event.ProviderEventID,
		&event.PaymentID,
		&event.EventType,
		&event.Payload,
		&event.CreatedAt,
	)

	if isUniqueViolation(err) {
		return ErrPaymentEventAlreadyProcessed
	}

	if err != nil {
		return fmt.Errorf("create payment event: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *postgresRepository) GetByProviderPaymentID(ctx context.Context, provider string, providerPaymentID string) (*Payment, error) {
	const query = `
        SELECT
            id,
            order_id,
            idempotency_key,
            provider,
            method,
            provider_payment_id,
            status,
            amount,
            currency,
            payment_url,
            qr_code_url,
            failure_reason,
            paid_at,
            failed_at,
            cancelled_at,
            expired_at,
            created_at,
            updated_at
        FROM payments
        WHERE provider = $1
            AND provider_payment_id = $2
    `

	var payment Payment

	err := r.db.QueryRow(
		ctx,
		query,
		provider,
		providerPaymentID,
	).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.IdempotencyKey,
		&payment.Provider,
		&payment.Method,
		&payment.ProviderPaymentID,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.PaymentURL,
		&payment.QRCodeURL,
		&payment.FailureReason,
		&payment.PaidAt,
		&payment.FailedAt,
		&payment.CancelledAt,
		&payment.ExpiredAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get payment by provider payment id: %w", err)
	}

	return &payment, nil
}
