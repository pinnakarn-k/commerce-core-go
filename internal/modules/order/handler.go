package order

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

// List godoc
// @Summary List orders
// @Description List current user orders.
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} OrdersSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /me/orders [get]
func (h *Handler) List(c *gin.Context) {
	authUser, err := authcontext.Get(c)
	if err != nil {
		response.FromError(c, err)
		return
	}

	orders, err := h.service.List(c.Request.Context(), authUser.ID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, toOrderResponses(orders))
}

// GetDetailByID godoc
// @Summary Get order detail
// @Description Get current user order detail by ID.
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} OrderDetailSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /me/orders/{id} [get]
func (h *Handler) GetDetailByID(c *gin.Context) {
	authUser, err := authcontext.Get(c)
	if err != nil {
		response.FromError(c, err)
		return
	}

	var p getOrderParams
	if err := c.ShouldBindUri(&p); err != nil {
		fields := validator.ParseValidationError(err, p)
		response.ValidationError(c, fields)
		return
	}

	detail, err := h.service.GetDetailByID(c.Request.Context(), authUser.ID, p.OrderID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, toOrderDetailResponse(*detail))
}
