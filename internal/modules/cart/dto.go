package cart

import "time"

type UpsertCartItemRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

type CartItemResponse struct {
	ID int64 `json:"id"`

	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`

	Quantity   int  `json:"quantity"`
	IsSelected bool `json:"is_selected"`

	Status  string `json:"status"`
	OrderID *int64 `json:"order_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toCartItemResponse(item CartItem) CartItemResponse {
	return CartItemResponse(item)
}

func toCartItemResponses(items []CartItem) []CartItemResponse {
	res := make([]CartItemResponse, 0, len(items))

	for _, item := range items {
		res = append(res, toCartItemResponse(item))
	}

	return res
}

type removeCartItemParams struct {
	ProductID int64 `uri:"product_id" binding:"required"`
}

type CartItemSuccessResponse struct {
	Data CartItemResponse `json:"data"`
}

type CartItemsSuccessResponse struct {
	Data []CartItemResponse `json:"data"`
}

type RemoveCartItemResponse struct {
	Message string `json:"message"`
}

type RemoveCartItemSuccessResponse struct {
	Data RemoveCartItemResponse `json:"data"`
}
