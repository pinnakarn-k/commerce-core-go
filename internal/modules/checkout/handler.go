package checkout

import (
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/authcontext"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/response"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/validator"

	"github.com/gin-gonic/gin"
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

// Checkout godoc
// @Summary Checkout cart
// @Description Create order and payment from selected cart items.
// @Tags Checkout
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CheckoutRequest true "Checkout request"
// @Success 200 {object} CheckoutSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /me/orders/checkout [post]
func (h *Handler) Checkout(c *gin.Context) {
	var req CheckoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fields := validator.ParseValidationError(err, req)
		response.ValidationError(c, fields)
		return
	}

	authUser, err := authcontext.Get(c)
	if err != nil {
		response.FromError(c, err)
		return
	}

	cmd := CheckoutCommand{
		UserID:          authUser.ID,
		IdempotencyKey:  req.IdempotencyKey,
		PaymentProvider: req.PaymentProvider,
		PaymentMethod:   req.PaymentMethod,
	}

	result, err := h.service.Checkout(c.Request.Context(), cmd)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, toCheckoutResponse(*result))
}
