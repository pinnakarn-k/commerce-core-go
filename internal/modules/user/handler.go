package user

import (
	"errors"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/authcontext"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/response"
	"github.com/pinnakarn-k/commerce-core-go/internal/platform/validator"

	"github.com/gin-gonic/gin"
)

var ErrNilService = errors.New("user handler: service is nil")

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

// Create godoc
// @Summary Create user
// @Description Register new user account.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "Create user request"
// @Success 200 {object} UserSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /users [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fields := validator.ParseValidationError(err, req)
		response.ValidationError(c, fields)
		return
	}

	cmd := CreateUserCommand(req)

	user, err := h.service.Create(c.Request.Context(), cmd)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, toUserResponse(*user))
}

// Me godoc
// @Summary Current user
// @Description Get current authenticated user.
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserSuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /me [get]
func (h *Handler) Me(c *gin.Context) {
	authUser, err := authcontext.Get(c)
	if err != nil {
		response.FromError(c, apperror.Unauthorized(
			"UNAUTHORIZED",
			"unauthorized",
			err,
		))
		return
	}

	user, err := h.service.Me(c.Request.Context(), authUser.ID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, toUserResponse(*user))
}
