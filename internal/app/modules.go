package app

import (
	"fmt"

	"github.com/pinnakarn-k/commerce-core-go/internal/modules/auth"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/cart"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/checkout"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/order"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/payment"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/product"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/user"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/database/postgres"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/httpmiddleware"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/token"

	"github.com/gin-gonic/gin"

	grpcserver "github.com/pinnakarn-k/commerce-core-go/internal/grpc"
)

func (a *App) registerModules(router gin.IRouter, db *postgres.Postgres) error {
	tokenMaker, err := token.NewJWTMaker(
		a.cfg.JWTSecret,
		a.cfg.JWTAccessTokenTTL,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitTokenMaker, err)
	}

	authMiddleware := httpmiddleware.RequireAuth(tokenMaker)

	userModule, err := user.NewModule(db.Pool)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitUserModule, err)
	}

	authModule, err := auth.NewModule(userModule.Repository(), tokenMaker)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitAuthModule, err)
	}

	cartModule, err := cart.NewModule(db.Pool)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitCartModule, err)
	}

	productRepo, err := product.NewRepository(db.Pool)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitProductModule, err)
	}

	orderModule, err := order.NewModule(db.Pool)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitOrderModule, err)
	}

	orderGRPCHandler := order.NewGRPCHandler(orderModule.Service())

	a.grpcServer = grpcserver.NewServer(orderGRPCHandler)

	paymentModule, err := payment.NewModule(db.Pool, orderModule.Repository())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitPaymentModule, err)
	}

	checkoutModule, err := checkout.NewModule(
		db.Pool,
		orderModule.Repository(),
		cartModule.Repository(),
		productRepo,
		paymentModule.Repository(),
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitCheckoutModule, err)
	}

	authModule.RegisterRoutes(router)
	userModule.RegisterRoutes(router, authMiddleware)
	cartModule.RegisterRoutes(router, authMiddleware)
	orderModule.RegisterRoutes(router, authMiddleware)
	checkoutModule.RegisterRoutes(router, authMiddleware)
	paymentModule.RegisterRoutes(router)

	return nil
}
