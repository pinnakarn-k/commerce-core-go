package product

import (
	"errors"
)

var (
	ErrNilDB              = errors.New("product repository: db is nil")
	ErrProductUnavailable = errors.New("product unavaliable")
)
