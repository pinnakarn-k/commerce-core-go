package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *Handler
}

func NewModule(userRepo UserRepository, tokenMaker TokenMaker) (*Module, error) {
	service, err := NewService(userRepo, tokenMaker)
	if err != nil {
		return nil, fmt.Errorf("init auth service: %w", err)
	}

	handler, err := NewHandler(service)
	if err != nil {
		return nil, fmt.Errorf("init auth handler: %w", err)
	}

	return &Module{
		handler: handler,
	}, nil
}

func (m *Module) RegisterRoutes(r gin.IRouter) {
	RegisterRoutes(r, m.handler)
}
