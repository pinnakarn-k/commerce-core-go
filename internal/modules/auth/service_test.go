package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/user"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindActiveUserByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	u, _ := args.Get(0).(*user.User)
	return u, args.Error(1)
}

type MockTokenMaker struct {
	mock.Mock
}

func (m *MockTokenMaker) CreateAccessToken(userID int64, role user.Role) (string, int64, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Get(1).(int64), args.Error(2)
}

func TestNewService_NilUserRepository(t *testing.T) {
	tokenMaker := new(MockTokenMaker)

	svc, err := NewService(nil, tokenMaker)

	require.Nil(t, svc)
	require.ErrorIs(t, err, ErrUserRepository)
}

func TestNewService_NilTokenMaker(t *testing.T) {
	userRepo := new(MockUserRepository)

	svc, err := NewService(userRepo, nil)

	require.Nil(t, svc)
	require.ErrorIs(t, err, ErrTokenMaker)
}

func TestNewService_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenMaker := new(MockTokenMaker)

	svc, err := NewService(userRepo, tokenMaker)

	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestService_Login_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenMaker := new(MockTokenMaker)

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("password123"),
		bcrypt.DefaultCost,
	)
	require.NoError(t, err)

	u := &user.User{
		ID:           1,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
		Role:         user.RoleUser,
		Status:       user.StatusActive,
	}

	userRepo.
		On("FindActiveUserByEmail", mock.Anything, "test@example.com").
		Return(u, nil)

	tokenMaker.
		On("CreateAccessToken", int64(1), user.RoleUser).
		Return("access-token-001", int64(3600), nil)

	svc := &Service{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}

	got, err := svc.Login(context.Background(), LoginCommand{
		Email:    "  TEST@EXAMPLE.COM  ",
		Password: "password123",
	})

	require.NoError(t, err)
	require.Equal(t, "access-token-001", got.AccessToken)
	require.Equal(t, "Bearer", got.TokenType)
	require.Equal(t, int64(3600), got.ExpiresIn)

	userRepo.AssertExpectations(t)
	tokenMaker.AssertExpectations(t)
}

func TestService_Login_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenMaker := new(MockTokenMaker)

	userRepo.
		On("FindActiveUserByEmail", mock.Anything, "test@example.com").
		Return((*user.User)(nil), user.ErrUserNotFound)

	svc := &Service{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}

	_, err := svc.Login(context.Background(), LoginCommand{
		Email:    "test@example.com",
		Password: "password123",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INVALID_CREDENTIALS", appErr.Code)

	userRepo.AssertExpectations(t)
	tokenMaker.AssertNotCalled(t, "CreateAccessToken")
}

func TestService_Login_UserRepositoryInternalError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenMaker := new(MockTokenMaker)

	userRepo.
		On("FindActiveUserByEmail", mock.Anything, "test@example.com").
		Return((*user.User)(nil), errors.New("database down"))

	svc := &Service{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}

	_, err := svc.Login(context.Background(), LoginCommand{
		Email:    "test@example.com",
		Password: "password123",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INTERNAL_SERVER_ERROR", appErr.Code)

	userRepo.AssertExpectations(t)
	tokenMaker.AssertNotCalled(t, "CreateAccessToken")
}

func TestService_Login_UserInactive(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenMaker := new(MockTokenMaker)

	u := &user.User{
		ID:           1,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		Role:         user.RoleUser,
		Status:       user.StatusDisabled,
	}

	userRepo.
		On("FindActiveUserByEmail", mock.Anything, "test@example.com").
		Return(u, nil)

	svc := &Service{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}

	_, err := svc.Login(context.Background(), LoginCommand{
		Email:    "test@example.com",
		Password: "password123",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "USER_INACTIVE", appErr.Code)

	userRepo.AssertExpectations(t)
	tokenMaker.AssertNotCalled(t, "CreateAccessToken")
}

func TestService_Login_InvalidPassword(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenMaker := new(MockTokenMaker)

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("correct-password"),
		bcrypt.DefaultCost,
	)
	require.NoError(t, err)

	u := &user.User{
		ID:           1,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
		Role:         user.RoleUser,
		Status:       user.StatusActive,
	}

	userRepo.
		On("FindActiveUserByEmail", mock.Anything, "test@example.com").
		Return(u, nil)

	svc := &Service{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}

	_, err = svc.Login(context.Background(), LoginCommand{
		Email:    "test@example.com",
		Password: "wrong-password",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INVALID_CREDENTIALS", appErr.Code)

	userRepo.AssertExpectations(t)
	tokenMaker.AssertNotCalled(t, "CreateAccessToken")
}

func TestService_Login_TokenMakerInternalError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenMaker := new(MockTokenMaker)

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("password123"),
		bcrypt.DefaultCost,
	)
	require.NoError(t, err)

	u := &user.User{
		ID:           1,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
		Role:         user.RoleUser,
		Status:       user.StatusActive,
	}

	userRepo.
		On("FindActiveUserByEmail", mock.Anything, "test@example.com").
		Return(u, nil)

	tokenMaker.
		On("CreateAccessToken", int64(1), user.RoleUser).
		Return("", int64(0), errors.New("sign token failed"))

	svc := &Service{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}

	_, err = svc.Login(context.Background(), LoginCommand{
		Email:    "test@example.com",
		Password: "password123",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INTERNAL_SERVER_ERROR", appErr.Code)

	userRepo.AssertExpectations(t)
	tokenMaker.AssertExpectations(t)
}
