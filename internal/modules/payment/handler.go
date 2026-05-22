package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/response"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/validator"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) (*Handler, error) {
	if service == nil {
		return nil, ErrNilService
	}

	return &Handler{
		service: service,
	}, nil
}

// MockWebhook godoc
// @Summary Mock payment webhook
// @Description Simulate payment provider webhook event.
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body MockWebhookRequest true "Mock webhook request"
// @Success 200 {object} PaymentSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /webhooks/payments/mock [post]
func (h *Handler) MockWebhook(c *gin.Context) {
	var req MockWebhookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fields := validator.ParseValidationError(err, req)
		response.ValidationError(c, fields)
		return
	}

	cmd := MockWebhookCommand(req)

	payment, err := h.service.HandleMockWebhook(c.Request.Context(), cmd)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, ToPaymentResponse(*payment))
}
