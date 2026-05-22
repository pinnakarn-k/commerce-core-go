package checkout

import (
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/order"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/payment"
)

type CheckoutRequest struct {
	IdempotencyKey  string `json:"idempotency_key" binding:"required"`
	PaymentProvider string `json:"payment_provider" binding:"required"`
	PaymentMethod   string `json:"payment_method" binding:"required"`
}

type CheckoutResponse struct {
	Order   order.OrderResponse     `json:"order"`
	Payment payment.PaymentResponse `json:"payment"`
}

func toCheckoutResponse(result CheckoutResult) CheckoutResponse {
	return CheckoutResponse{
		Order:   order.ToOrderResponse(result.Order),
		Payment: payment.ToPaymentResponse(result.Payment),
	}
}

type CheckoutSuccessResponse struct {
	Data CheckoutResponse `json:"data"`
}
