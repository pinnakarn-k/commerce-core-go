package order

type CreateOrderItemInput struct {
	CartItemID int64
	ProductID  int64
	Quantity   int

	ProductName string
	ProductSKU  string

	UnitPriceAmount int
	Currency        string

	LineTotalAmount int
}
