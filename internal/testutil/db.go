package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func NewTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	_ = godotenv.Load("configs/.env.test")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("skip integration test: DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)

	err = db.Ping(context.Background())
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TruncateTables(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(
		context.Background(),
		`
		TRUNCATE TABLE
			payment_events,
			payments,
			cart_items,
			order_items,
			orders,
			products,
			users
		RESTART IDENTITY CASCADE
		`,
	)
	require.NoError(t, err)
}
