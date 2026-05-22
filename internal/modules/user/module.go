package user

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	repo    Repository
	handler *Handler
}

func (m *Module) Repository() Repository {
	return m.repo
}

func NewModule(pool *pgxpool.Pool) (*Module, error) {
	repo, err := NewRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("init user repository: %w", err)
	}

	service, err := NewService(repo)
	if err != nil {
		return nil, fmt.Errorf("init user service: %w", err)
	}

	handler, err := NewHandler(service)
	if err != nil {
		return nil, fmt.Errorf("init user handler: %w", err)
	}

	return &Module{
		repo:    repo,
		handler: handler,
	}, nil
}

func (m *Module) RegisterRoutes(
	r gin.IRouter,
	authMiddleware gin.HandlerFunc,
) {
	RegisterRoutes(r, m.handler, authMiddleware)
}
