package cart

import "time"

type CartItem struct {
	ID int64

	UserID    int64
	ProductID int64

	Quantity   int
	IsSelected bool

	Status  string
	OrderID *int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type CheckoutItem struct {
	CartItemID int64
	ProductID  int64
	Quantity   int

	ProductName string
	ProductSKU  string

	UnitPriceAmount int
	Currency        string

	LineTotalAmount int
}
