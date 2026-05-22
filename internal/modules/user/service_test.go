package user

import (
	"context"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*User)
	return u, args.Error(1)
}

func (m *MockRepository) FindActiveUserByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	u, _ := args.Get(0).(*User)
	return u, args.Error(1)
}

func TestService_Create(t *testing.T) {
	repo := new(MockRepository)

	repo.
		On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*User)
			u.ID = 1
		}).
		Return(nil)

	svc, err := NewService(repo)
	require.NoError(t, err)

	cmd := CreateUserCommand{
		Name:     "User A",
		Email:    "a@example.com ",
		Password: "123456",
	}

	user, err := svc.Create(context.Background(), cmd)
	require.NoError(t, err)

	require.Equal(t, int64(1), user.ID)
	require.Equal(t, "User A", user.Name)
	require.Equal(t, "a@example.com", user.Email)
	require.NotEmpty(t, user.PasswordHash)

	repo.AssertExpectations(t)
}

func TestService_Create_EmailAlreadyExists(t *testing.T) {
	repo := new(MockRepository)

	repo.
		On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
		Return(ErrEmailAlreadyExists)

	svc, err := NewService(repo)
	require.NoError(t, err)

	cmd := CreateUserCommand{
		Name:     "User A",
		Email:    "a@example.com",
		Password: "123456",
	}

	_, err = svc.Create(context.Background(), cmd)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.KindConflict, appErr.Kind)
	require.Equal(t, "EMAIL_ALREADY_EXISTS", appErr.Code)

	repo.AssertExpectations(t)
}

func TestService_Me(t *testing.T) {
	repo := new(MockRepository)

	repo.
		On("GetByID", mock.Anything, int64(1)).
		Return(&User{
			ID:    1,
			Name:  "User A",
			Email: "a@example.com",
			Role:  RoleUser,
		}, nil)

	svc, err := NewService(repo)
	require.NoError(t, err)

	user, err := svc.Me(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, int64(1), user.ID)
	require.Equal(t, "User A", user.Name)
	require.Equal(t, "a@example.com", user.Email)
	require.Equal(t, RoleUser, user.Role)

	repo.AssertExpectations(t)
}

func TestService_Me_UserNotFound(t *testing.T) {
	repo := new(MockRepository)

	repo.
		On("GetByID", mock.Anything, int64(1)).
		Return((*User)(nil), ErrUserNotFound)

	svc, err := NewService(repo)
	require.NoError(t, err)

	_, err = svc.Me(context.Background(), 1)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.KindNotFound, appErr.Kind)
	require.Equal(t, "USER_NOT_FOUND", appErr.Code)

	repo.AssertExpectations(t)
}
