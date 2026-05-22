package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/user"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrRequestJWTSecretKey     = errors.New("jwt secret key is required")
	ErrInvalidJWTTTL           = errors.New("jwt ttl must be greater than zero")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
)

type JWTMaker struct {
	secretKey string
	ttl       time.Duration
}

type Claims struct {
	UserID int64     `json:"user_id"`
	Role   user.Role `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTMaker(secretKey string, ttl time.Duration) (*JWTMaker, error) {
	if secretKey == "" {
		return nil, ErrRequestJWTSecretKey
	}

	if ttl <= 0 {
		return nil, ErrInvalidJWTTTL
	}

	return &JWTMaker{
		secretKey: secretKey,
		ttl:       ttl,
	}, nil
}

func (m *JWTMaker) CreateAccessToken(userID int64, role user.Role) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", 0, err
	}

	return signedToken, int64(m.ttl.Seconds()), nil
}

func (m *JWTMaker) VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
			}

			return []byte(m.secretKey), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
