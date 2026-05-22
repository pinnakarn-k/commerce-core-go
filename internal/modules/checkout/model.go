package checkout

import (
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/order"
	"github.com/pinnakarn-k/commerce-core-go/internal/modules/payment"
)

type CheckoutResult struct {
	Order   order.Order
	Payment payment.Payment
}
