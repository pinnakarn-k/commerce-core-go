package testconfig

import (
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/config"
)

func Load(t *testing.T) config.Config {
	t.Helper()

	t.Setenv("SERVICE", "commerce-core-go")
	t.Setenv("ENV", "test")
	t.Setenv("APP_PORT", "0")
	t.Setenv("GRPC_PORT", "0")

	t.Setenv(
		"DATABASE_URL",
		"postgres://test:test@localhost:5433/commerce_test?sslmode=disable",
	)

	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_ACCESS_TOKEN_TTL_MINUTES", "60")

	return config.Load()
}
