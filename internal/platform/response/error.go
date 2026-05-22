package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/apperror"

	"github.com/gin-gonic/gin"
)

func Error(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.JSON(status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func ValidationError(c *gin.Context, fields []FieldError) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Fields:  fields,
		},
	})
}

func statusFromKind(kind apperror.ErrorKind) int {
	switch kind {
	case apperror.KindBadRequest:
		return http.StatusBadRequest
	case apperror.KindUnauthorized:
		return http.StatusUnauthorized
	case apperror.KindForbidden:
		return http.StatusForbidden
	case apperror.KindNotFound:
		return http.StatusNotFound
	case apperror.KindConflict:
		return http.StatusConflict
	case apperror.KindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func FromError(c *gin.Context, err error) {
	var appErr *apperror.AppError

	if errors.As(err, &appErr) {
		status := statusFromKind(appErr.Kind)

		if status >= http.StatusInternalServerError {
			slog.Error(
				"application error",
				"error", err,
				"code", appErr.Code,
				"kind", appErr.Kind,
				"status", status,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
		}

		Error(
			c,
			status,
			appErr.Code,
			appErr.Message,
		)

		return
	}

	slog.Error(
		"unexpected error",
		"error", err,
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	Error(
		c,
		http.StatusInternalServerError,
		"INTERNAL_SERVER_ERROR",
		"internal server error",
	)
}
