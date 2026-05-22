package payment

import (
	"context"
	"testing"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreatePayment(
	ctx context.Context,
	tx pgx.Tx,
	payment *Payment,
) error {
	args := m.Called(ctx, tx, payment)
	return args.Error(0)
}

func (m *MockRepository) FindByOrderIDAndIdempotencyKey(
	ctx context.Context,
	orderID int64,
	idempotencyKey string,
) (*Payment, error) {
	args := m.Called(ctx, orderID, idempotencyKey)
	payment, _ := args.Get(0).(*Payment)
	return payment, args.Error(1)
}

func (m *MockRepository) GetByID(
	ctx context.Context,
	id int64,
) (*Payment, error) {
	args := m.Called(ctx, id)
	payment, _ := args.Get(0).(*Payment)
	return payment, args.Error(1)
}

func (m *MockRepository) GetByProviderPaymentID(
	ctx context.Context,
	provider string,
	providerPaymentID string,
) (*Payment, error) {
	args := m.Called(ctx, provider, providerPaymentID)
	payment, _ := args.Get(0).(*Payment)
	return payment, args.Error(1)
}

func (m *MockRepository) MarkSucceeded(
	ctx context.Context,
	tx pgx.Tx,
	id int64,
) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepository) MarkOrderPaid(
	ctx context.Context,
	tx pgx.Tx,
	orderID int64,
) error {
	args := m.Called(ctx, tx, orderID)
	return args.Error(0)
}

func (m *MockRepository) MarkFailed(
	ctx context.Context,
	tx pgx.Tx,
	id int64,
	reason string,
) error {
	args := m.Called(ctx, tx, id, reason)
	return args.Error(0)
}

func (m *MockRepository) MarkCancelled(
	ctx context.Context,
	tx pgx.Tx,
	id int64,
) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepository) MarkExpired(
	ctx context.Context,
	tx pgx.Tx,
	id int64,
) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepository) CreateEvent(
	ctx context.Context,
	tx pgx.Tx,
	event *PaymentEvent,
) error {
	args := m.Called(ctx, tx, event)
	return args.Error(0)
}

func TestStatus_IsFinal(t *testing.T) {
	require.False(t, PaymentStatusPending.IsFinal())
	require.True(t, PaymentStatusSucceeded.IsFinal())
	require.True(t, PaymentStatusFailed.IsFinal())
	require.True(t, PaymentStatusCancelled.IsFinal())
	require.True(t, PaymentStatusExpired.IsFinal())
}

func TestStatus_Valid(t *testing.T) {
	require.True(t, PaymentStatusPending.Valid())
	require.True(t, PaymentStatusSucceeded.Valid())
	require.True(t, PaymentStatusFailed.Valid())
	require.True(t, PaymentStatusCancelled.Valid())
	require.True(t, PaymentStatusExpired.Valid())
	require.False(t, Status("unknown").Valid())
}

func TestService_HandleMockWebhook_InvalidPaymentEventID(t *testing.T) {
	svc := &Service{}

	_, err := svc.HandleMockWebhook(context.Background(), MockWebhookCommand{
		EventID:           " ",
		Provider:          "mock",
		ProviderPaymentID: "mock_pay_1",
		Status:            "succeeded",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INVALID_PAYMENT_EVENT_ID", appErr.Code)
}

func TestService_HandleMockWebhook_InvalidPaymentProvider(t *testing.T) {
	svc := &Service{}

	_, err := svc.HandleMockWebhook(context.Background(), MockWebhookCommand{
		EventID:           "evt-001",
		Provider:          " ",
		ProviderPaymentID: "mock_pay_1",
		Status:            "succeeded",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INVALID_PAYMENT_PROVIDER", appErr.Code)
}

func TestService_HandleMockWebhook_InvalidProviderPaymentID(t *testing.T) {
	svc := &Service{}

	_, err := svc.HandleMockWebhook(context.Background(), MockWebhookCommand{
		EventID:           "evt-001",
		Provider:          "mock",
		ProviderPaymentID: " ",
		Status:            "succeeded",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "INVALID_PROVIDER_PAYMENT_ID", appErr.Code)
}

func TestService_HandleMockWebhook_PaymentNotFound(t *testing.T) {
	repo := new(MockRepository)

	repo.
		On(
			"GetByProviderPaymentID",
			mock.Anything,
			"mock",
			"mock_pay_1",
		).
		Return((*Payment)(nil), ErrPaymentNotFound)

	svc := &Service{
		repo: repo,
	}

	_, err := svc.HandleMockWebhook(context.Background(), MockWebhookCommand{
		EventID:           "evt-001",
		Provider:          "mock",
		ProviderPaymentID: "mock_pay_1",
		Status:            "succeeded",
	})

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "PAYMENT_NOT_FOUND", appErr.Code)

	repo.AssertExpectations(t)
}

func TestService_HandleMockWebhook_FinalPayment_ReturnsExistingPayment(t *testing.T) {
	repo := new(MockRepository)

	existingPayment := &Payment{
		ID:                1,
		OrderID:           10,
		IdempotencyKey:    "idem-001",
		Provider:          "mock",
		Method:            "promptpay",
		ProviderPaymentID: "mock_pay_10",
		Status:            PaymentStatusSucceeded,
		Amount:            20000,
		Currency:          "THB",
	}

	repo.
		On(
			"GetByProviderPaymentID",
			mock.Anything,
			"mock",
			"mock_pay_10",
		).
		Return(existingPayment, nil)

	svc := &Service{
		repo: repo,
		db:   nil,
	}

	got, err := svc.HandleMockWebhook(context.Background(), MockWebhookCommand{
		EventID:           "evt-002",
		Provider:          "mock",
		ProviderPaymentID: "mock_pay_10",
		Status:            "failed",
	})

	require.NoError(t, err)
	require.Equal(t, existingPayment, got)
	require.Equal(t, PaymentStatusSucceeded, got.Status)

	repo.AssertExpectations(t)
}
