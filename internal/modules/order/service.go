package order

import (
	"context"
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo Repository
	db   *pgxpool.Pool
}

func NewService(
	repo Repository,
	db *pgxpool.Pool,
) (*Service, error) {
	if repo == nil {
		return nil, ErrNilRepository
	}
	if db == nil {
		return nil, ErrNilDB
	}

	return &Service{
		repo: repo,
		db:   db,
	}, nil
}

func (s *Service) List(ctx context.Context, userID int64) ([]Order, error) {
	if userID <= 0 {
		return nil, InvalidUserID(nil)
	}

	orders, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	return orders, nil
}

func (s *Service) GetDetailByID(ctx context.Context, userID int64, orderID int64) (*OrderDetail, error) {
	if userID <= 0 {
		return nil, InvalidUserID(nil)
	}
	if orderID <= 0 {
		return nil, InvalidOrderID(nil)
	}

	detail, err := s.repo.GetDetailByID(ctx, userID, orderID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			return nil, OrderNotFound(err)
		}

		return nil, apperror.Internal(err)
	}

	return detail, nil
}
