package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/auth"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testapp"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHTTP_Login_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	application := testapp.New(t)

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("0123456789"),
		bcrypt.MinCost,
	)
	require.NoError(t, err)

	_, err = db.Exec(
		t.Context(),
		`
		INSERT INTO users (
			name,
			email,
			password_hash,
			role,
			status
		)
		VALUES (
			'Test User',
			'login@example.com',
			$1,
			'user',
			'active'
		)
		`,
		string(passwordHash),
	)
	require.NoError(t, err)

	body, err := json.Marshal(map[string]string{
		"email":    "login@example.com",
		"password": "0123456789",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	application.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var res auth.LoginSuccessResponse

	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	require.NotEmpty(t, res.Data.AccessToken)
	require.Equal(t, "Bearer", res.Data.TokenType)
	require.Greater(t, res.Data.ExpiresIn, int64(0))
}

func TestHTTP_Login_InvalidPassword(t *testing.T) {
	db := testutil.NewTestDB(t)
	testutil.TruncateTables(t, db)

	application := testapp.New(t)

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("0123456789"),
		bcrypt.MinCost,
	)
	require.NoError(t, err)

	_, err = db.Exec(
		t.Context(),
		`
		INSERT INTO users (
			name,
			email,
			password_hash,
			role,
			status
		)
		VALUES (
			'Test User',
			'invalid-password@example.com',
			$1,
			'user',
			'active'
		)
		`,
		string(passwordHash),
	)
	require.NoError(t, err)

	body, err := json.Marshal(map[string]string{
		"email":    "invalid-password@example.com",
		"password": "wrong-password",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		bytes.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	application.Handler().ServeHTTP(rec, req)

	require.Equal(
		t,
		http.StatusUnauthorized,
		rec.Code,
	)
}

func TestHTTP_Login_ValidationError(t *testing.T) {
	application := testapp.New(t)

	body, err := json.Marshal(map[string]string{
		"email": "not-an-email",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		bytes.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	application.Handler().ServeHTTP(rec, req)

	require.Equal(
		t,
		http.StatusBadRequest,
		rec.Code,
	)
}
