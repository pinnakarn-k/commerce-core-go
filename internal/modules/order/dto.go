package order

import (
	"time"
)

type getOrderParams struct {
	OrderID int64 `uri:"id" binding:"required"`
}

type OrderResponse struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	IdempotencyKey string     `json:"idempotency_key"`
	Status         string     `json:"status"`
	TotalAmount    int        `json:"total_amount"`
	Currency       string     `json:"currency"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
}

func ToOrderResponse(order Order) OrderResponse {
	return OrderResponse{
		ID:             order.ID,
		UserID:         order.UserID,
		IdempotencyKey: order.IdempotencyKey,
		Status:         string(order.Status),
		TotalAmount:    order.TotalAmount,
		Currency:       order.Currency,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
		PaidAt:         order.PaidAt,
		CancelledAt:    order.CancelledAt,
	}
}

func toOrderResponses(orders []Order) []OrderResponse {
	res := make([]OrderResponse, 0, len(orders))

	for _, o := range orders {
		res = append(res, ToOrderResponse(o))
	}

	return res
}

type OrderItemResponse struct {
	ID              int64     `json:"id"`
	OrderID         int64     `json:"order_id"`
	ProductID       int64     `json:"product_id"`
	ProductName     string    `json:"product_name"`
	ProductSKU      string    `json:"product_sku"`
	Quantity        int       `json:"quantity"`
	UnitPriceAmount int       `json:"unit_price_amount"`
	LineTotalAmount int       `json:"line_total_amount"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
}

func toOrderItemResponse(item OrderItem) OrderItemResponse {
	return OrderItemResponse(item)
}

func toOrderItemResponses(items []OrderItem) []OrderItemResponse {
	res := make([]OrderItemResponse, 0, len(items))

	for _, item := range items {
		res = append(res, toOrderItemResponse(item))
	}

	return res
}

type OrderDetailResponse struct {
	Order OrderResponse       `json:"order"`
	Items []OrderItemResponse `json:"items"`
}

func toOrderDetailResponse(detail OrderDetail) OrderDetailResponse {
	return OrderDetailResponse{
		Order: ToOrderResponse(detail.Order),
		Items: toOrderItemResponses(detail.Items),
	}
}

type OrdersSuccessResponse struct {
	Data []OrderResponse `json:"data"`
}

type OrderDetailSuccessResponse struct {
	Data OrderDetailResponse `json:"data"`
}
