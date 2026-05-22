package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) UpsertItem(ctx context.Context, item *CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartRepository) ListItems(ctx context.Context, userID int64) ([]CartItem, error) {
	args := m.Called(ctx, userID)
	items, _ := args.Get(0).([]CartItem)
	return items, args.Error(1)
}

func (m *MockCartRepository) RemoveItem(ctx context.Context, userID int64, productID int64) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

func TestNewService_NilRepository(t *testing.T) {
	svc, err := NewService(nil)

	require.Nil(t, svc)
	require.ErrorIs(t, err, ErrNilRepository)
}

func TestNewService_Success(t *testing.T) {
	repo := new(MockCartRepository)

	svc, err := NewService(repo)

	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestService_UpsertItem_Success(t *testing.T) {
	repo := new(MockCartRepository)

	repo.
		On("UpsertItem", mock.Anything, mock.MatchedBy(func(item *CartItem) bool {
			return item.UserID == 1 &&
				item.ProductID == 10 &&
				item.Quantity == 2
		})).
		Return(nil)

	svc := &Service{
		repo: repo,
	}

	got, err := svc.UpsertItem(context.Background(), UpsertItemCommand{
		UserID:    1,
		ProductID: 10,
		Quantity:  2,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), got.UserID)
	require.Equal(t, int64(10), got.ProductID)
	require.Equal(t, 2, got.Quantity)

	repo.AssertExpectations(t)
}

func TestService_UpsertItem_ProductUnavailable(t *testing.T) {
	repo := new(MockCartRepository)

	repo.
		On("UpsertItem", mock.Anything, mock.AnythingOfType("*cart.CartItem")).
		Return(ErrProductUnavailable)

	svc := &Service{
		repo: repo,
	}

	_, err := svc.UpsertItem(context.Background(), UpsertItemCommand{
		UserID:    1,
		ProductID: 10,
		Quantity:  2,
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "PRODUCT_UNAVAILABLE", appErr.Code)

	repo.AssertExpectations(t)
}

func TestService_UpsertItem_InternalError(t *testing.T) {
	repo := new(MockCartRepository)

	repo.
		On("UpsertItem", mock.Anything, mock.AnythingOfType("*cart.CartItem")).
		Return(errors.New("database down"))

	svc := &Service{
		repo: repo,
	}

	_, err := svc.UpsertItem(context.Background(), UpsertItemCommand{
		UserID:    1,
		ProductID: 10,
		Quantity:  2,
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INTERNAL_SERVER_ERROR", appErr.Code)

	repo.AssertExpectations(t)
}

func TestService_ListItems_Success(t *testing.T) {
	repo := new(MockCartRepository)

	items := []CartItem{
		{
			ID:        1,
			UserID:    1,
			ProductID: 10,
			Quantity:  2,
		},
		{
			ID:        2,
			UserID:    1,
			ProductID: 20,
			Quantity:  1,
		},
	}

	repo.
		On("ListItems", mock.Anything, int64(1)).
		Return(items, nil)

	svc := &Service{
		repo: repo,
	}

	got, err := svc.ListItems(context.Background(), ListItemsCommand{
		UserID: 1,
	})

	require.NoError(t, err)
	require.Equal(t, items, got)

	repo.AssertExpectations(t)
}

func TestService_ListItems_InternalError(t *testing.T) {
	repo := new(MockCartRepository)

	repo.
		On("ListItems", mock.Anything, int64(1)).
		Return([]CartItem(nil), errors.New("database down"))

	svc := &Service{
		repo: repo,
	}

	_, err := svc.ListItems(context.Background(), ListItemsCommand{
		UserID: 1,
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INTERNAL_SERVER_ERROR", appErr.Code)

	repo.AssertExpectations(t)
}

func TestService_RemoveItem_Success(t *testing.T) {
	repo := new(MockCartRepository)

	repo.
		On("RemoveItem", mock.Anything, int64(1), int64(10)).
		Return(nil)

	svc := &Service{
		repo: repo,
	}

	err := svc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID:    1,
		ProductID: 10,
	})

	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestService_RemoveItem_CartItemNotFound(t *testing.T) {
	repo := new(MockCartRepository)

	repo.
		On("RemoveItem", mock.Anything, int64(1), int64(10)).
		Return(ErrCartItemNotFound)

	svc := &Service{
		repo: repo,
	}

	err := svc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID:    1,
		ProductID: 10,
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "CART_ITEM_NOT_FOUND", appErr.Code)

	repo.AssertExpectations(t)
}

func TestService_RemoveItem_InternalError(t *testing.T) {
	repo := new(MockCartRepository)

	repo.
		On("RemoveItem", mock.Anything, int64(1), int64(10)).
		Return(errors.New("database down"))

	svc := &Service{
		repo: repo,
	}

	err := svc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID:    1,
		ProductID: 10,
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INTERNAL_SERVER_ERROR", appErr.Code)

	repo.AssertExpectations(t)
}
