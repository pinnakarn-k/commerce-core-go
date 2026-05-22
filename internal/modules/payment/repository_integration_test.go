package payment

import (
	"context"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/testutil"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testpayment"

	"github.com/stretchr/testify/require"
)

func TestRepository_CreatePayment_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	paymentURL := "https://payment.local/pay"
	qrCodeURL := "https://payment.local/qr"

	p := &Payment{
		OrderID:           orderID,
		IdempotencyKey:    "idem-001",
		Provider:          "mock",
		Method:            "promptpay",
		ProviderPaymentID: "mock_pay_001",
		Status:            PaymentStatusPending,
		Amount:            2000,
		Currency:          "THB",
		PaymentURL:        &paymentURL,
		QRCodeURL:         &qrCodeURL,
	}

	err = repo.CreatePayment(ctx, tx, p)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	require.NotZero(t, p.ID)
	require.Equal(t, orderID, p.OrderID)
	require.Equal(t, "idem-001", p.IdempotencyKey)
	require.Equal(t, "mock", p.Provider)
	require.Equal(t, "promptpay", p.Method)
	require.Equal(t, "mock_pay_001", p.ProviderPaymentID)
	require.Equal(t, PaymentStatusPending, p.Status)
	require.Equal(t, 2000, p.Amount)
	require.Equal(t, "THB", p.Currency)

	require.NotNil(t, p.PaymentURL)
	require.Equal(t, paymentURL, *p.PaymentURL)

	require.NotNil(t, p.QRCodeURL)
	require.Equal(t, qrCodeURL, *p.QRCodeURL)

	require.Nil(t, p.FailureReason)
	require.Nil(t, p.PaidAt)
	require.Nil(t, p.FailedAt)
	require.Nil(t, p.CancelledAt)
	require.Nil(t, p.ExpiredAt)

	require.False(t, p.CreatedAt.IsZero())
	require.False(t, p.UpdatedAt.IsZero())
}

func TestRepository_FindByOrderIDAndIdempotencyKey_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})
	require.NoError(t, err)

	payment, err := repo.FindByOrderIDAndIdempotencyKey(
		ctx,
		orderID,
		"idem-001",
	)

	require.NoError(t, err)

	require.Equal(t, paymentID, payment.ID)
	require.Equal(t, orderID, payment.OrderID)
	require.Equal(t, "idem-001", payment.IdempotencyKey)
	require.Equal(t, "mock", payment.Provider)
	require.Equal(t, "promptpay", payment.Method)
	require.Equal(t, "mock_pay_001", payment.ProviderPaymentID)
	require.Equal(t, PaymentStatusPending, payment.Status)
	require.Equal(t, 2000, payment.Amount)
	require.Equal(t, "THB", payment.Currency)
}

func TestRepository_FindByOrderIDAndIdempotencyKey_NotFound(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	payment, err := repo.FindByOrderIDAndIdempotencyKey(
		ctx,
		int64(999),
		"idem-not-found",
	)

	require.Nil(t, payment)
	require.ErrorIs(t, err, ErrPaymentNotFound)
}

func TestRepository_GetByID_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})
	require.NoError(t, err)

	payment, err := repo.GetByID(ctx, paymentID)

	require.NoError(t, err)
	require.Equal(t, paymentID, payment.ID)
	require.Equal(t, orderID, payment.OrderID)
	require.Equal(t, "idem-001", payment.IdempotencyKey)
	require.Equal(t, "mock_pay_001", payment.ProviderPaymentID)
	require.Equal(t, PaymentStatusPending, payment.Status)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	payment, err := repo.GetByID(ctx, int64(999))

	require.Nil(t, payment)
	require.ErrorIs(t, err, ErrPaymentNotFound)
}

func TestRepository_CreateEvent_Duplicate(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	event1 := &PaymentEvent{
		Provider:        "mock",
		ProviderEventID: "evt-001",
		PaymentID:       paymentID,
		EventType:       "payment.succeeded",
		Payload:         []byte(`{"status":"succeeded"}`),
	}

	err = repo.CreateEvent(ctx, tx, event1)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	tx2, err := db.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		_ = tx2.Rollback(ctx)
	}()

	event2 := &PaymentEvent{
		Provider:        "mock",
		ProviderEventID: "evt-001",
		PaymentID:       paymentID,
		EventType:       "payment.succeeded",
		Payload:         []byte(`{"status":"succeeded"}`),
	}

	err = repo.CreateEvent(ctx, tx2, event2)

	require.ErrorIs(t, err, ErrPaymentEventAlreadyProcessed)
}

func TestRepository_CreateEvent_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	event := &PaymentEvent{
		Provider:        "mock",
		ProviderEventID: "evt-001",
		PaymentID:       paymentID,
		EventType:       "payment.succeeded",
		Payload:         []byte(`{"status":"succeeded"}`),
	}

	err = repo.CreateEvent(ctx, tx, event)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	require.NotZero(t, event.ID)
	require.Equal(t, "mock", event.Provider)
	require.Equal(t, "evt-001", event.ProviderEventID)
	require.Equal(t, paymentID, event.PaymentID)
	require.Equal(t, "payment.succeeded", event.EventType)
	require.JSONEq(t, `{"status":"succeeded"}`, string(event.Payload))
	require.False(t, event.CreatedAt.IsZero())
}

func TestRepository_MarkSucceeded_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	err = repo.MarkSucceeded(ctx, tx, paymentID)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	var status Status
	var paidAt *string

	err = db.QueryRow(
		ctx,
		`
		SELECT status, paid_at::text
		FROM payments
		WHERE id = $1
		`,
		paymentID,
	).Scan(&status, &paidAt)
	require.NoError(t, err)

	require.Equal(t, PaymentStatusSucceeded, status)
	require.NotNil(t, paidAt)
}

func TestRepository_MarkOrderPaid_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	err = repo.MarkOrderPaid(ctx, tx, orderID)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	var status string
	var paidAt *string

	err = db.QueryRow(
		ctx,
		`
		SELECT status, paid_at::text
		FROM orders
		WHERE id = $1
		`,
		orderID,
	).Scan(&status, &paidAt)
	require.NoError(t, err)

	require.Equal(t, "paid", status)
	require.NotNil(t, paidAt)
}

func TestRepository_MarkFailed_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	err = repo.MarkFailed(ctx, tx, paymentID, "card declined")
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	var status Status
	var failureReason *string
	var failedAt *string

	err = db.QueryRow(ctx, `
		SELECT status, failure_reason, failed_at::text
		FROM payments
		WHERE id = $1
	`, paymentID).Scan(&status, &failureReason, &failedAt)
	require.NoError(t, err)

	require.Equal(t, PaymentStatusFailed, status)
	require.NotNil(t, failureReason)
	require.Equal(t, "card declined", *failureReason)
	require.NotNil(t, failedAt)
}

func TestRepository_MarkCancelled_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	err = repo.MarkCancelled(ctx, tx, paymentID)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	var status Status
	var cancelledAt *string

	err = db.QueryRow(ctx, `
		SELECT status, cancelled_at::text
		FROM payments
		WHERE id = $1
	`, paymentID).Scan(&status, &cancelledAt)
	require.NoError(t, err)

	require.Equal(t, PaymentStatusCancelled, status)
	require.NotNil(t, cancelledAt)
}

func TestRepository_MarkExpired_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})
	require.NoError(t, err)

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	err = repo.MarkExpired(ctx, tx, paymentID)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	var status Status
	var expiredAt *string

	err = db.QueryRow(ctx, `
		SELECT status, expired_at::text
		FROM payments
		WHERE id = $1
	`, paymentID).Scan(&status, &expiredAt)
	require.NoError(t, err)

	require.Equal(t, PaymentStatusExpired, status)
	require.NotNil(t, expiredAt)
}

func TestRepository_GetByProviderPaymentID_Success(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(t, db)

	orderID := testutil.CreateTestOrder(t, db, userID)

	paymentID := testpayment.CreatePayment(t, db, testpayment.CreatePaymentInput{
		OrderID: orderID,
		Amount:  2000,
	})
	require.NoError(t, err)

	got, err := repo.GetByProviderPaymentID(ctx, "mock", "mock_pay_001")

	require.NoError(t, err)
	require.Equal(t, paymentID, got.ID)
	require.Equal(t, orderID, got.OrderID)
	require.Equal(t, "mock", got.Provider)
	require.Equal(t, "mock_pay_001", got.ProviderPaymentID)
	require.Equal(t, PaymentStatusPending, got.Status)
}

func TestRepository_GetByProviderPaymentID_NotFound(t *testing.T) {
	ctx := context.Background()

	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	got, err := repo.GetByProviderPaymentID(
		ctx,
		"mock",
		"mock_pay_not_found",
	)

	require.Nil(t, got)
	require.ErrorIs(t, err, ErrPaymentNotFound)
}
