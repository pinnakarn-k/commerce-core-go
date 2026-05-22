package cart

import (
	"context"
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
)

type CartRepository interface {
	ListItems(ctx context.Context, userID int64) ([]CartItem, error)
	UpsertItem(ctx context.Context, item *CartItem) error
	RemoveItem(ctx context.Context, userID int64, productID int64) error
}

type Service struct {
	repo CartRepository
}

func NewService(repo CartRepository) (*Service, error) {
	if repo == nil {
		return nil, ErrNilRepository
	}

	return &Service{
		repo: repo,
	}, nil
}

func (s *Service) UpsertItem(ctx context.Context, cmd UpsertItemCommand) (*CartItem, error) {
	item := &CartItem{
		UserID:    cmd.UserID,
		ProductID: cmd.ProductID,
		Quantity:  cmd.Quantity,
	}

	if err := s.repo.UpsertItem(ctx, item); err != nil {
		if errors.Is(err, ErrProductUnavailable) {
			return nil, ProductUnavailable(err)
		}

		return nil, apperror.Internal(err)
	}

	return item, nil
}

func (s *Service) ListItems(ctx context.Context, cmd ListItemsCommand) ([]CartItem, error) {
	items, err := s.repo.ListItems(ctx, cmd.UserID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	return items, nil
}

func (s *Service) RemoveItem(ctx context.Context, cmd RemoveItemCommand) error {
	if err := s.repo.RemoveItem(ctx, cmd.UserID, cmd.ProductID); err != nil {
		if errors.Is(err, ErrCartItemNotFound) {
			return CartItemNotFound(err)
		}

		return apperror.Internal(err)
	}

	return nil
}
