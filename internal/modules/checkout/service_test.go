package checkout

import (
	"context"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/cart"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/order"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/payment"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) FindOrderByIdempotencyKey(ctx context.Context, userID int64, key string) (*order.Order, error) {
	args := m.Called(ctx, userID, key)
	o, _ := args.Get(0).(*order.Order)
	return o, args.Error(1)
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, tx pgx.Tx, o *order.Order) error {
	args := m.Called(ctx, tx, o)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateOrderItem(ctx context.Context, tx pgx.Tx, orderID int64, input order.CreateOrderItemInput) error {
	args := m.Called(ctx, tx, orderID, input)
	return args.Error(0)
}

type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) ListCheckoutItems(ctx context.Context, userID int64) ([]cart.CheckoutItem, error) {
	args := m.Called(ctx, userID)
	items, _ := args.Get(0).([]cart.CheckoutItem)
	return items, args.Error(1)
}

func (m *MockCartRepository) MarkItemsPurchased(ctx context.Context, tx pgx.Tx, userID int64, orderID int64) error {
	args := m.Called(ctx, tx, userID, orderID)
	return args.Error(0)
}

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) CreatePayment(ctx context.Context, tx pgx.Tx, p *payment.Payment) error {
	args := m.Called(ctx, tx, p)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByOrderIDAndIdempotencyKey(ctx context.Context, orderID int64, key string) (*payment.Payment, error) {
	args := m.Called(ctx, orderID, key)
	p, _ := args.Get(0).(*payment.Payment)
	return p, args.Error(1)
}

func TestService_Checkout_IdempotencyHit(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	paymentRepo := new(MockPaymentRepository)

	existingOrder := &order.Order{
		ID:             10,
		UserID:         1,
		IdempotencyKey: "idem-001",
		Status:         order.OrderStatusPending,
		TotalAmount:    20000,
		Currency:       "THB",
	}

	existingPayment := &payment.Payment{
		ID:                5,
		OrderID:           10,
		IdempotencyKey:    "idem-001",
		Provider:          "mock",
		Method:            "promptpay",
		ProviderPaymentID: "mock_pay_10",
		Status:            payment.PaymentStatusPending,
		Amount:            20000,
		Currency:          "THB",
	}

	orderRepo.
		On("FindOrderByIdempotencyKey", mock.Anything, int64(1), "idem-001").
		Return(existingOrder, nil)

	paymentRepo.
		On("FindByOrderIDAndIdempotencyKey", mock.Anything, int64(10), "idem-001").
		Return(existingPayment, nil)

	svc := &Service{
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
	}

	got, err := svc.Checkout(context.Background(), CheckoutCommand{
		UserID:          1,
		IdempotencyKey:  "idem-001",
		PaymentProvider: "mock",
		PaymentMethod:   "promptpay",
	})

	require.NoError(t, err)
	require.Equal(t, *existingOrder, got.Order)
	require.Equal(t, *existingPayment, got.Payment)
}

func TestService_Checkout_CartEmpty(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	cartRepo := new(MockCartRepository)

	orderRepo.
		On("FindOrderByIdempotencyKey", mock.Anything, int64(1), "idem-001").
		Return((*order.Order)(nil), order.ErrOrderNotFound)

	cartRepo.
		On("ListCheckoutItems", mock.Anything, int64(1)).
		Return([]cart.CheckoutItem{}, nil)

	svc := &Service{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
	}

	_, err := svc.Checkout(context.Background(), CheckoutCommand{
		UserID:          1,
		IdempotencyKey:  "idem-001",
		PaymentProvider: "mock",
		PaymentMethod:   "promptpay",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "CART_EMPTY", appErr.Code)
}
