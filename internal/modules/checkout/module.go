package checkout

import (
	"fmt"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/cart"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/order"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/payment"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/product"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	handler *Handler
}

func NewModule(
	pool *pgxpool.Pool,
	orderRepo order.Repository,
	cartRepo cart.Repository,
	productRepo product.Repository,
	paymentRepo payment.Repository,
) (*Module, error) {
	service, err := NewService(orderRepo, cartRepo, productRepo, paymentRepo, pool)
	if err != nil {
		return nil, fmt.Errorf("init checkout service: %w", err)
	}

	handler, err := NewHandler(service)
	if err != nil {
		return nil, fmt.Errorf("init checkout handler: %w", err)
	}

	return &Module{
		handler: handler,
	}, nil
}

func (m *Module) RegisterRoutes(
	r gin.IRouter,
	authMiddleware gin.HandlerFunc,
) {
	RegisterRoutes(r, m.handler, authMiddleware)
}
