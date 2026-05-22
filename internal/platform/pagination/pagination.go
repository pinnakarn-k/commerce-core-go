package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

type Params struct {
	Page   int
	Limit  int
	Offset int
}

func FromGin(c *gin.Context) Params {
	page := parseInt(c.Query("page"), DefaultPage)
	limit := parseInt(c.Query("limit"), DefaultLimit)

	if page < 1 {
		page = DefaultPage
	}

	if limit < 1 {
		limit = DefaultLimit
	}

	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Params{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func TotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}

	return int((total + int64(limit) - 1) / int64(limit))
}

func parseInt(v string, fallback int) int {
	if v == "" {
		return fallback
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return i
}
