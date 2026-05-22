package testapp

import (
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/app"
	"github.com/pinnakarn-k/commerce-core-go/internal/testutil/testconfig"

	"github.com/stretchr/testify/require"
)

func New(t *testing.T) *app.App {
	t.Helper()

	cfg := testconfig.Load(t)

	application, err := app.New(cfg)
	require.NoError(t, err)

	return application
}
