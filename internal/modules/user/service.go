package user

import (
	"context"
	"errors"
	"strings"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"golang.org/x/crypto/bcrypt"
)

var ErrNilRepository = errors.New("user service: repository is nil")

type Service struct {
	repo Repository
}

func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, ErrNilRepository
	}

	return &Service{
		repo: repo,
	}, nil
}

func (s *Service) Create(ctx context.Context, cmd CreateUserCommand) (*User, error) {
	name := strings.TrimSpace(cmd.Name)
	email := strings.ToLower(strings.TrimSpace(cmd.Email))

	hashed, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	user := &User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashed),
		Role:         RoleUser,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return nil, EmailAlreadyExists(err)
		}

		return nil, apperror.Internal(err)
	}

	return user, nil
}

func (s *Service) Me(ctx context.Context, userID int64) (*User, error) {
	if userID <= 0 {
		return nil, InvalidUserID(nil)
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, UserNotFound(err)
		}

		return nil, apperror.Internal(err)
	}

	return user, nil
}
