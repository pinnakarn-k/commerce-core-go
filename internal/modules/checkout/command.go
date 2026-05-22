package checkout

type CheckoutCommand struct {
	UserID int64

	IdempotencyKey string

	PaymentProvider string
	PaymentMethod   string
}
