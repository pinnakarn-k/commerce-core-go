package checkout

import (
	"context"
	"fmt"
	"strings"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/cart"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/order"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/payment"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/product"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, tx pgx.Tx, payment *payment.Payment) error
	FindByOrderIDAndIdempotencyKey(ctx context.Context, orderID int64, idempotencyKey string) (*payment.Payment, error)
}

type CartRepository interface {
	ListCheckoutItems(ctx context.Context, userID int64) ([]cart.CheckoutItem, error)
	MarkItemsPurchased(ctx context.Context, tx pgx.Tx, userID int64, orderID int64) error
}

type OrderRepository interface {
	FindOrderByIdempotencyKey(ctx context.Context, userID int64, idempotencyKey string) (*order.Order, error)
	CreateOrder(ctx context.Context, tx pgx.Tx, order *order.Order) error
	CreateOrderItem(ctx context.Context, tx pgx.Tx, orderID int64, input order.CreateOrderItemInput) error
}

type Service struct {
	orderRepo   OrderRepository
	cartRepo    CartRepository
	productRepo product.Repository
	paymentRepo PaymentRepository
	db          *pgxpool.Pool
}

func NewService(
	orderRepo OrderRepository,
	cartRepo CartRepository,
	productRepo product.Repository,
	paymentRepo PaymentRepository,
	db *pgxpool.Pool,
) (*Service, error) {
	if orderRepo == nil {
		return nil, ErrNilRepository
	}
	if cartRepo == nil {
		return nil, ErrNilCartRepository
	}
	if productRepo == nil {
		return nil, ErrNilProductRepository
	}
	if paymentRepo == nil {
		return nil, ErrNilPaymentRepository
	}
	if db == nil {
		return nil, ErrNilDB
	}

	return &Service{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		paymentRepo: paymentRepo,
		db:          db,
	}, nil
}

func (s *Service) Checkout(ctx context.Context, cmd CheckoutCommand) (*CheckoutResult, error) {
	// 1. validation
	userID := cmd.UserID
	idempotencyKey := strings.TrimSpace(cmd.IdempotencyKey)
	paymentProvider := strings.TrimSpace(cmd.PaymentProvider)
	paymentMethod := strings.TrimSpace(cmd.PaymentMethod)

	if userID <= 0 {
		return nil, InvalidUserID(nil)
	}

	if idempotencyKey == "" {
		return nil, InvalidIdempotencyKey(nil)
	}

	if paymentProvider == "" {
		return nil, InvalidPaymentProvider(nil)
	}

	if paymentMethod == "" {
		return nil, InvalidPaymentMethod(nil)
	}

	// 2. idempotency
	existing, err := s.orderRepo.FindOrderByIdempotencyKey(ctx, userID, idempotencyKey)
	if err == nil {
		existingPayment, err := s.paymentRepo.FindByOrderIDAndIdempotencyKey(
			ctx,
			existing.ID,
			idempotencyKey,
		)
		if err != nil {
			return nil, apperror.Internal(err)
		}

		return &CheckoutResult{
			Order:   *existing,
			Payment: *existingPayment,
		}, nil
	}

	// 3. get snapshot
	items, err := s.cartRepo.ListCheckoutItems(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if len(items) == 0 {
		return nil, CartEmpty(nil)
	}

	// 4. begin tx
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 5. deduct stock (loop)
	for _, item := range items {
		if err := s.productRepo.DeductStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
			return nil, ProductUnavailable(err)
		}
	}

	// 6. calculate total
	total := 0
	for _, item := range items {
		total += item.LineTotalAmount
	}

	// 7. create order
	createdOrder := &order.Order{
		UserID:         userID,
		IdempotencyKey: idempotencyKey,
		Status:         order.OrderStatusPending,
		TotalAmount:    total,
		Currency:       "THB",
	}

	if err := s.orderRepo.CreateOrder(ctx, tx, createdOrder); err != nil {
		return nil, apperror.Internal(err)
	}

	// 8. insert order_items
	for _, item := range items {
		input := order.CreateOrderItemInput{
			ProductID:       item.ProductID,
			ProductName:     item.ProductName,
			ProductSKU:      item.ProductSKU,
			Quantity:        item.Quantity,
			UnitPriceAmount: item.UnitPriceAmount,
			LineTotalAmount: item.LineTotalAmount,
			Currency:        item.Currency,
		}

		if err := s.orderRepo.CreateOrderItem(ctx, tx, createdOrder.ID, input); err != nil {
			return nil, apperror.Internal(err)
		}
	}

	// 9. mark cart_items purchased
	if err := s.cartRepo.MarkItemsPurchased(ctx, tx, userID, createdOrder.ID); err != nil {
		return nil, apperror.Internal(err)
	}

	// 10. create payment
	mockProviderPaymentID := fmt.Sprintf("mock_pay_%d", createdOrder.ID)
	mockPaymentURL := "https://mock-payment.local/payments/pay"
	mockQRCodeURL := "https://mock-payment.local/payments/qr"

	createdPayment := &payment.Payment{
		OrderID:           createdOrder.ID,
		IdempotencyKey:    idempotencyKey,
		Provider:          paymentProvider,
		Method:            paymentMethod,
		ProviderPaymentID: mockProviderPaymentID,
		Status:            payment.PaymentStatusPending,
		Amount:            createdOrder.TotalAmount,
		Currency:          createdOrder.Currency,
		PaymentURL:        stringPtr(mockPaymentURL),
		QRCodeURL:         stringPtr(mockQRCodeURL),
	}
	if err := s.paymentRepo.CreatePayment(ctx, tx, createdPayment); err != nil {
		return nil, apperror.Internal(err)
	}

	// 11. commit
	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal(err)
	}

	return &CheckoutResult{
		Order:   *createdOrder,
		Payment: *createdPayment,
	}, nil
}

func stringPtr(value string) *string {
	return &value
}
