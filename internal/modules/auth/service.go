package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/user"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"golang.org/x/crypto/bcrypt"
)

type TokenMaker interface {
	CreateAccessToken(userID int64, role user.Role) (token string, expiresIn int64, err error)
}

type UserRepository interface {
	FindActiveUserByEmail(ctx context.Context, email string) (*user.User, error)
}

type Service struct {
	userRepo   UserRepository
	tokenMaker TokenMaker
}

func NewService(userRepo UserRepository, tokenMaker TokenMaker) (*Service, error) {
	if userRepo == nil {
		return nil, ErrUserRepository
	}

	if tokenMaker == nil {
		return nil, ErrTokenMaker
	}

	return &Service{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}, nil
}

func (s *Service) Login(ctx context.Context, cmd LoginCommand) (*LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(cmd.Email))

	u, err := s.userRepo.FindActiveUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, InvalidCredentials(err)
		}

		return nil, apperror.Internal(err)
	}

	if u.Status != user.StatusActive {
		return nil, UserInactive(nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(cmd.Password)); err != nil {
		return nil, InvalidCredentials(err)
	}

	accessToken, expiresIn, err := s.tokenMaker.CreateAccessToken(u.ID, u.Role)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	return &LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}
