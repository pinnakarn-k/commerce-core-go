package testauth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/app"

	"github.com/stretchr/testify/require"
)

func LoginAndGetToken(
	t *testing.T,
	application *app.App,
	email string,
	password string,
) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
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

	require.Equal(t, http.StatusOK, rec.Code)

	var res struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	require.NotEmpty(t, res.Data.AccessToken)

	return res.Data.AccessToken
}
