package order

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	repo    Repository
	service *Service
	handler *Handler
}

func (m *Module) Repository() Repository {
	return m.repo
}

func (m *Module) Service() *Service {
	return m.service
}

func NewModule(pool *pgxpool.Pool) (*Module, error) {
	repo, err := NewRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("init order repository: %w", err)
	}

	service, err := NewService(repo, pool)
	if err != nil {
		return nil, fmt.Errorf("init order service: %w", err)
	}

	handler, err := NewHandler(service)
	if err != nil {
		return nil, fmt.Errorf("init order handler: %w", err)
	}

	return &Module{
		repo:    repo,
		service: service,
		handler: handler,
	}, nil
}

func (m *Module) RegisterRoutes(
	r gin.IRouter,
	authMiddleware gin.HandlerFunc,
) {
	RegisterRoutes(r, m.handler, authMiddleware)
}
