package response

import (
	"net/http"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/pagination"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data: data,
	})
}

func OKWithMeta(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data: data,
		Meta: meta,
	})
}

func calcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}

	return int((total + int64(limit) - 1) / int64(limit))
}

func Paginated(
	c *gin.Context,
	data any,
	params pagination.Params,
	total int64,
) {
	OKWithMeta(c, data, PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: calcTotalPages(total, params.Limit),
	})
}
