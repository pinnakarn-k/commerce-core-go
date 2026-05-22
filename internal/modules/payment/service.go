package payment

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookRepository interface {
	GetByProviderPaymentID(ctx context.Context, provider string, providerPaymentID string) (*Payment, error)
	CreateEvent(ctx context.Context, tx pgx.Tx, event *PaymentEvent) error
	MarkSucceeded(ctx context.Context, tx pgx.Tx, id int64) error
	MarkOrderPaid(ctx context.Context, tx pgx.Tx, orderID int64) error
	MarkFailed(ctx context.Context, tx pgx.Tx, id int64, reason string) error
	MarkCancelled(ctx context.Context, tx pgx.Tx, id int64) error
	MarkExpired(ctx context.Context, tx pgx.Tx, id int64) error
}

type Service struct {
	repo WebhookRepository
	db   *pgxpool.Pool
}

func NewService(repo Repository, db *pgxpool.Pool) (*Service, error) {
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

func (s *Service) HandleMockWebhook(ctx context.Context, cmd MockWebhookCommand) (*Payment, error) {
	eventID := strings.TrimSpace(cmd.EventID)
	provider := strings.TrimSpace(cmd.Provider)
	providerPaymentID := strings.TrimSpace(cmd.ProviderPaymentID)
	status := Status(strings.TrimSpace(cmd.Status))

	if eventID == "" {
		return nil, InvalidPaymentEventID(nil)
	}

	if provider == "" {
		return nil, InvalidPaymentProvider(nil)
	}

	if providerPaymentID == "" {
		return nil, InvalidProviderPaymentID(nil)
	}

	payment, err := s.repo.GetByProviderPaymentID(
		ctx,
		provider,
		providerPaymentID,
	)
	if errors.Is(err, ErrPaymentNotFound) {
		return nil, PaymentNotFound(err)
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}

	if payment.Status.IsFinal() {
		return payment, nil
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	event := &PaymentEvent{
		Provider:        payment.Provider,
		ProviderEventID: eventID,
		PaymentID:       payment.ID,
		EventType:       string(status),
		Payload:         payload,
	}

	if err := s.repo.CreateEvent(ctx, tx, event); err != nil {
		if errors.Is(err, ErrPaymentEventAlreadyProcessed) {
			return payment, nil
		}

		return nil, apperror.Internal(err)
	}

	switch status {
	case PaymentStatusSucceeded:
		if err := s.repo.MarkSucceeded(ctx, tx, payment.ID); err != nil {
			return nil, apperror.Internal(err)
		}

		if err := s.repo.MarkOrderPaid(ctx, tx, payment.OrderID); err != nil {
			return nil, apperror.Internal(err)
		}

		payment.Status = PaymentStatusSucceeded

	case PaymentStatusFailed:
		reason := strings.TrimSpace(cmd.Reason)
		if reason == "" {
			reason = "payment failed"
		}

		if err := s.repo.MarkFailed(ctx, tx, payment.ID, reason); err != nil {
			return nil, apperror.Internal(err)
		}

		payment.Status = PaymentStatusFailed
		payment.FailureReason = &reason

	case PaymentStatusCancelled:
		if err := s.repo.MarkCancelled(ctx, tx, payment.ID); err != nil {
			return nil, apperror.Internal(err)
		}

		payment.Status = PaymentStatusCancelled

	case PaymentStatusExpired:
		if err := s.repo.MarkExpired(ctx, tx, payment.ID); err != nil {
			return nil, apperror.Internal(err)
		}

		payment.Status = PaymentStatusExpired

	default:
		return nil, InvalidPaymentStatus(nil)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperror.Internal(err)
	}

	return payment, nil
}
