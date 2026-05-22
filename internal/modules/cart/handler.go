package cart

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

// UpsertItem godoc
// @Summary Add or update cart item
// @Description Add product to cart or update quantity.
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpsertCartItemRequest true "Cart item request"
// @Success 200 {object} CartItemSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /me/cart/items [put]
func (h *Handler) UpsertItem(c *gin.Context) {
	authUser, err := authcontext.Get(c)
	if err != nil {
		response.FromError(c, err)
		return
	}

	var req UpsertCartItemRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fields := validator.ParseValidationError(err, req)
		response.ValidationError(c, fields)
		return
	}

	cmd := UpsertItemCommand{
		UserID:    authUser.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}

	item, err := h.service.UpsertItem(c.Request.Context(), cmd)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, toCartItemResponse(*item))
}

// ListItems godoc
// @Summary List cart items
// @Description List active cart items for current user.
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Success 200 {object} CartItemsSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /me/cart [get]
func (h *Handler) ListItems(c *gin.Context) {
	authUser, err := authcontext.Get(c)
	if err != nil {
		response.FromError(c, err)
		return
	}

	cmd := ListItemsCommand{
		UserID: authUser.ID,
	}

	items, err := h.service.ListItems(c.Request.Context(), cmd)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, toCartItemResponses(items))
}

// RemoveItem godoc
// @Summary Remove cart item
// @Description Remove product from cart.
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Param product_id path int true "Product ID"
// @Success 200 {object} RemoveCartItemSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /me/cart/items/{product_id} [delete]
func (h *Handler) RemoveItem(c *gin.Context) {
	authUser, err := authcontext.Get(c)
	if err != nil {
		response.FromError(c, err)
		return
	}

	var p removeCartItemParams
	if err := c.ShouldBindUri(&p); err != nil {
		fields := validator.ParseValidationError(err, p)
		response.ValidationError(c, fields)
		return
	}

	cmd := RemoveItemCommand{
		UserID:    authUser.ID,
		ProductID: p.ProductID,
	}

	if err := h.service.RemoveItem(c.Request.Context(), cmd); err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, gin.H{
		"message": "cart item removed",
	})
}
