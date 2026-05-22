package testuser

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func CreateActiveUser(
	t *testing.T,
	db *pgxpool.Pool,
	email string,
	password string,
) int64 {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.MinCost,
	)
	require.NoError(t, err)

	var userID int64
	err = db.QueryRow(
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
			$1,
			$2,
			'user',
			'active'
		)
		RETURNING id
		`,
		email,
		string(passwordHash),
	).Scan(&userID)
	require.NoError(t, err)

	return userID
}
