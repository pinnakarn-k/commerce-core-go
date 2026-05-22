package auth

import (
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

// Login godoc
// @Summary Login
// @Description Authenticate user and return JWT access token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} LoginSuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fields := validator.ParseValidationError(err, req)
		response.ValidationError(c, fields)
		return
	}

	cmd := LoginCommand(req)

	res, err := h.service.Login(c.Request.Context(), cmd)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, res)
}
