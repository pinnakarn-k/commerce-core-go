package payment

import "time"

type MockWebhookRequest struct {
	EventID           string `json:"event_id" binding:"required"`
	Provider          string `json:"provider" binding:"required"`
	ProviderPaymentID string `json:"provider_payment_id" binding:"required"`
	Status            string `json:"status" binding:"required"`
	Reason            string `json:"reason"`
}

type PaymentResponse struct {
	ID                int64      `json:"id"`
	OrderID           int64      `json:"order_id"`
	IdempotencyKey    string     `json:"idempotency_key"`
	Provider          string     `json:"provider"`
	Method            string     `json:"method"`
	ProviderPaymentID string     `json:"provider_payment_id,omitempty"`
	Status            string     `json:"status"`
	Amount            int        `json:"amount"`
	Currency          string     `json:"currency"`
	PaymentURL        *string    `json:"payment_url,omitempty"`
	QRCodeURL         *string    `json:"qr_code_url,omitempty"`
	FailureReason     *string    `json:"failure_reason,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	FailedAt          *time.Time `json:"failed_at,omitempty"`
	CancelledAt       *time.Time `json:"cancelled_at,omitempty"`
	ExpiredAt         *time.Time `json:"expired_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func ToPaymentResponse(payment Payment) PaymentResponse {
	return PaymentResponse{
		ID:                payment.ID,
		OrderID:           payment.OrderID,
		IdempotencyKey:    payment.IdempotencyKey,
		Provider:          payment.Provider,
		Method:            payment.Method,
		ProviderPaymentID: payment.ProviderPaymentID,
		Status:            string(payment.Status),
		Amount:            payment.Amount,
		Currency:          payment.Currency,
		PaymentURL:        payment.PaymentURL,
		QRCodeURL:         payment.QRCodeURL,
		FailureReason:     payment.FailureReason,
		PaidAt:            payment.PaidAt,
		FailedAt:          payment.FailedAt,
		CancelledAt:       payment.CancelledAt,
		ExpiredAt:         payment.ExpiredAt,
		CreatedAt:         payment.CreatedAt,
		UpdatedAt:         payment.UpdatedAt,
	}
}

type PaymentSuccessResponse struct {
	Data PaymentResponse `json:"data"`
}
