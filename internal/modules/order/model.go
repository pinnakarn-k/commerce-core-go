package order

import (
	"time"
)

type Status string

const (
	OrderStatusPending   Status = "pending"
	OrderStatusPaid      Status = "paid"
	OrderStatusCancelled Status = "cancelled"
)

func (status Status) IsFinal() bool {
	return status == OrderStatusPaid || status == OrderStatusCancelled
}

func (status Status) Valid() bool {
	switch status {
	case OrderStatusPending, OrderStatusPaid, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

type Order struct {
	ID int64

	UserID int64

	IdempotencyKey string

	Status Status

	TotalAmount int
	Currency    string

	CreatedAt time.Time
	UpdatedAt time.Time

	PaidAt      *time.Time
	CancelledAt *time.Time
}

type OrderItem struct {
	ID int64

	OrderID   int64
	ProductID int64

	ProductName string
	ProductSKU  string

	Quantity int

	UnitPriceAmount int
	LineTotalAmount int
	Currency        string

	CreatedAt time.Time
}

type OrderDetail struct {
	Order Order
	Items []OrderItem
}
