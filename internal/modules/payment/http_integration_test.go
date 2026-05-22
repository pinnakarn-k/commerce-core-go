package payment_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/payment"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testapp"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testproduct"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testuser"

	"github.com/stretchr/testify/require"
)

func TestHTTP_MockWebhook_Succeeded(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	application := testapp.New(t)

	userID := testuser.CreateActiveUser(
		t,
		db,
		"webhook@example.com",
		"0123456789",
	)

	productID := testproduct.CreateActiveProduct(
		t,
		db,
		testproduct.CreateProductInput{
			Name:        "Keyboard",
			SKU:         "SKU-WEBHOOK",
			PriceAmount: 1500,
			StockQty:    10,
		},
	)

	var orderID int64

	err := db.QueryRow(
		t.Context(),
		`
		INSERT INTO orders (
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency
		)
		VALUES (
			$1,
			'order-webhook-001',
			'pending',
			3000,
			'THB'
		)
		RETURNING id
		`,
		userID,
	).Scan(&orderID)

	require.NoError(t, err)

	_ = productID

	var paymentID int64

	err = db.QueryRow(
		t.Context(),
		`
		INSERT INTO payments (
			order_id,
			idempotency_key,
			provider,
			method,
			provider_payment_id,
			status,
			amount,
			currency
		)
		VALUES (
			$1,
			'payment-webhook-001',
			'mock',
			'promptpay',
			'mock-pay-001',
			'pending',
			3000,
			'THB'
		)
		RETURNING id
		`,
		orderID,
	).Scan(&paymentID)

	require.NoError(t, err)

	payload := payment.MockWebhookRequest{
		EventID:           "evt-001",
		Provider:          "mock",
		ProviderPaymentID: "mock-pay-001",
		Status:            "succeeded",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/webhooks/payments/mock",
		bytes.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	application.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var paymentStatus string

	err = db.QueryRow(
		t.Context(),
		`
		SELECT status
		FROM payments
		WHERE id = $1
		`,
		paymentID,
	).Scan(&paymentStatus)

	require.NoError(t, err)

	require.Equal(t, "succeeded", paymentStatus)

	var orderStatus string

	err = db.QueryRow(
		t.Context(),
		`
		SELECT status
		FROM orders
		WHERE id = $1
		`,
		orderID,
	).Scan(&orderStatus)

	require.NoError(t, err)

	require.Equal(t, "paid", orderStatus)

	var eventCount int

	err = db.QueryRow(
		t.Context(),
		`
		SELECT COUNT(*)
		FROM payment_events
		WHERE provider_event_id = 'evt-001'
		`,
	).Scan(&eventCount)

	require.NoError(t, err)

	require.Equal(t, 1, eventCount)
}

func TestHTTP_MockWebhook_DuplicateEvent(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	application := testapp.New(t)

	userID := testuser.CreateActiveUser(
		t,
		db,
		"duplicate-webhook@example.com",
		"0123456789",
	)

	var orderID int64

	err := db.QueryRow(
		t.Context(),
		`
		INSERT INTO orders (
			user_id,
			idempotency_key,
			status,
			total_amount,
			currency
		)
		VALUES (
			$1,
			'order-duplicate-001',
			'pending',
			3000,
			'THB'
		)
		RETURNING id
		`,
		userID,
	).Scan(&orderID)

	require.NoError(t, err)

	var paymentID int64

	err = db.QueryRow(
		t.Context(),
		`
		INSERT INTO payments (
			order_id,
			idempotency_key,
			provider,
			method,
			provider_payment_id,
			status,
			amount,
			currency
		)
		VALUES (
			$1,
			'payment-duplicate-001',
			'mock',
			'promptpay',
			'mock-pay-duplicate',
			'pending',
			3000,
			'THB'
		)
		RETURNING id
		`,
		orderID,
	).Scan(&paymentID)

	require.NoError(t, err)

	payload := payment.MockWebhookRequest{
		EventID:           "evt-duplicate-001",
		Provider:          "mock",
		ProviderPaymentID: "mock-pay-duplicate",
		Status:            "succeeded",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	for range 2 {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/webhooks/payments/mock",
			bytes.NewReader(body),
		)

		req.Header.Set(
			"Content-Type",
			"application/json",
		)

		rec := httptest.NewRecorder()

		application.Handler().ServeHTTP(rec, req)
	}

	var eventCount int

	err = db.QueryRow(
		t.Context(),
		`
		SELECT COUNT(*)
		FROM payment_events
		WHERE provider_event_id = 'evt-duplicate-001'
		`,
	).Scan(&eventCount)

	require.NoError(t, err)

	require.Equal(t, 1, eventCount)

	var paymentStatus string

	err = db.QueryRow(
		t.Context(),
		`
		SELECT status
		FROM payments
		WHERE id = $1
		`,
		paymentID,
	).Scan(&paymentStatus)

	require.NoError(t, err)

	require.Equal(t, "succeeded", paymentStatus)
}
