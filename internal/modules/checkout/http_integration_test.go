package checkout_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/checkout"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testapp"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testauth"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testproduct"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testuser"

	"github.com/stretchr/testify/require"
)

func TestHTTP_Checkout_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	application := testapp.New(t)

	userID := testuser.CreateActiveUser(
		t,
		db,
		"checkout-http@example.com",
		"0123456789",
	)

	token := testauth.LoginAndGetToken(
		t,
		application,
		"checkout-http@example.com",
		"0123456789",
	)

	productID := testproduct.CreateActiveProduct(
		t,
		db,
		testproduct.CreateProductInput{
			Name:        "Keyboard",
			SKU:         "SKU-HTTP-CHECKOUT",
			PriceAmount: 1500,
			StockQty:    10,
		},
	)

	_, err := db.Exec(
		t.Context(),
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

	body, err := json.Marshal(map[string]string{
		"idempotency_key":  "checkout-http-001",
		"payment_provider": "mock",
		"payment_method":   "promptpay",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/me/orders/checkout",
		bytes.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	application.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var res checkout.CheckoutSuccessResponse

	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	require.Greater(t, res.Data.Order.ID, int64(0))
	require.Greater(
		t,
		res.Data.Payment.ID,
		int64(0),
	)
	require.Equal(
		t,
		"pending",
		res.Data.Payment.Status,
	)
}

func TestHTTP_Checkout_WithoutToken(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	application := testapp.New(t)

	checkoutReq := checkout.CheckoutRequest{
		IdempotencyKey:  "checkout-no-token",
		PaymentProvider: "mock",
		PaymentMethod:   "promptpay",
	}

	body, err := json.Marshal(checkoutReq)
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/me/orders/checkout",
		bytes.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	application.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
