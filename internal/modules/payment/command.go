package payment

type MockWebhookCommand struct {
	EventID           string
	Provider          string
	ProviderPaymentID string
	Status            string
	Reason            string
}
