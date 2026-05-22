package payment

import "time"

type Status string

const (
	PaymentStatusPending   Status = "pending"
	PaymentStatusSucceeded Status = "succeeded"
	PaymentStatusFailed    Status = "failed"
	PaymentStatusCancelled Status = "cancelled"
	PaymentStatusExpired   Status = "expired"
)

func (status Status) IsFinal() bool {
	return status == PaymentStatusSucceeded ||
		status == PaymentStatusFailed ||
		status == PaymentStatusCancelled ||
		status == PaymentStatusExpired
}

func (status Status) Valid() bool {
	switch status {
	case PaymentStatusPending,
		PaymentStatusSucceeded,
		PaymentStatusFailed,
		PaymentStatusCancelled,
		PaymentStatusExpired:
		return true
	default:
		return false
	}
}

type Payment struct {
	ID int64

	OrderID int64

	IdempotencyKey string

	Provider string
	Method   string

	ProviderPaymentID string

	Status Status

	Amount   int
	Currency string

	PaymentURL *string
	QRCodeURL  *string

	FailureReason *string

	PaidAt      *time.Time
	FailedAt    *time.Time
	CancelledAt *time.Time
	ExpiredAt   *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type PaymentEvent struct {
	ID int64

	Provider        string
	ProviderEventID string
	PaymentID       int64
	EventType       string
	Payload         []byte

	CreatedAt time.Time
}
