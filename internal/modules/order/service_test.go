package order

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) ListByUserID(
	ctx context.Context,
	userID int64,
) ([]Order, error) {
	args := m.Called(ctx, userID)
	orders, _ := args.Get(0).([]Order)
	return orders, args.Error(1)
}

func (m *MockOrderRepository) GetDetailByID(
	ctx context.Context,
	userID int64,
	orderID int64,
) (*OrderDetail, error) {
	args := m.Called(ctx, userID, orderID)
	detail, _ := args.Get(0).(*OrderDetail)
	return detail, args.Error(1)
}

func (m *MockOrderRepository) FindOrderByIdempotencyKey(
	ctx context.Context,
	userID int64,
	key string,
) (*Order, error) {
	args := m.Called(ctx, userID, key)
	order, _ := args.Get(0).(*Order)
	return order, args.Error(1)
}

func (m *MockOrderRepository) CreateOrder(
	ctx context.Context,
	tx pgx.Tx,
	order *Order,
) error {
	args := m.Called(ctx, tx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateOrderItem(
	ctx context.Context,
	tx pgx.Tx,
	orderID int64,
	input CreateOrderItemInput,
) error {
	args := m.Called(ctx, tx, orderID, input)
	return args.Error(0)
}

func newTestService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func TestService_List(t *testing.T) {
	repo := new(MockOrderRepository)

	orders := []Order{
		{
			ID:          1,
			UserID:      1,
			Status:      OrderStatusPending,
			TotalAmount: 20000,
			Currency:    "THB",
		},
	}

	repo.
		On("ListByUserID", mock.Anything, int64(1)).
		Return(orders, nil)

	svc := newTestService(repo)

	got, err := svc.List(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, int64(1), got[0].ID)

	repo.AssertExpectations(t)
}

func TestService_GetDetailByID(t *testing.T) {
	repo := new(MockOrderRepository)

	detail := &OrderDetail{
		Order: Order{
			ID:          1,
			UserID:      1,
			Status:      OrderStatusPending,
			TotalAmount: 20000,
			Currency:    "THB",
		},
		Items: []OrderItem{
			{
				ID:              1,
				OrderID:         1,
				ProductID:       1,
				ProductName:     "Wireless Mouse",
				ProductSKU:      "SKU-001",
				Quantity:        2,
				UnitPriceAmount: 10000,
				LineTotalAmount: 20000,
				Currency:        "THB",
			},
		},
	}

	repo.
		On("GetDetailByID", mock.Anything, int64(1), int64(1)).
		Return(detail, nil)

	svc := newTestService(repo)

	got, err := svc.GetDetailByID(context.Background(), 1, 1)

	require.NoError(t, err)
	require.Equal(t, int64(1), got.Order.ID)
	require.Len(t, got.Items, 1)
	require.Equal(t, "Wireless Mouse", got.Items[0].ProductName)

	repo.AssertExpectations(t)
}
