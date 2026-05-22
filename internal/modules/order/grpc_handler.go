package order

import (
	"context"
	"time"

	orderv1 "github.com/pinnakarn-k/commerce-core-go/internal/gen/order/v1"
)

type GRPCHandler struct {
	orderv1.UnimplementedOrderQueryServiceServer
	service *Service
}

func NewGRPCHandler(service *Service) *GRPCHandler {
	return &GRPCHandler{
		service: service,
	}
}

func (h *GRPCHandler) GetOrderDetail(
	ctx context.Context,
	req *orderv1.GetOrderDetailRequest,
) (*orderv1.OrderDetailResponse, error) {
	detail, err := h.service.GetDetailByID(
		ctx,
		req.GetUserId(),
		req.GetOrderId(),
	)
	if err != nil {
		return nil, err
	}

	return toGRPCOrderDetailResponse(*detail), nil
}

func toGRPCOrderDetailResponse(
	detail OrderDetail,
) *orderv1.OrderDetailResponse {
	items := make([]*orderv1.OrderItem, 0, len(detail.Items))

	for _, item := range detail.Items {
		items = append(items, toGRPCOrderItem(item))
	}

	return &orderv1.OrderDetailResponse{
		Order: toGRPCOrder(detail.Order),
		Items: items,
	}
}

func toGRPCOrder(order Order) *orderv1.Order {
	res := &orderv1.Order{
		Id:             order.ID,
		UserId:         order.UserID,
		IdempotencyKey: order.IdempotencyKey,
		Status:         string(order.Status),
		TotalAmount:    int32(order.TotalAmount),
		Currency:       order.Currency,
		CreatedAt:      order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      order.UpdatedAt.Format(time.RFC3339),
	}

	if order.PaidAt != nil {
		paidAt := order.PaidAt.Format(time.RFC3339)
		res.PaidAt = &paidAt
	}

	if order.CancelledAt != nil {
		cancelledAt := order.CancelledAt.Format(time.RFC3339)
		res.CancelledAt = &cancelledAt
	}

	return res
}

func toGRPCOrderItem(item OrderItem) *orderv1.OrderItem {
	return &orderv1.OrderItem{
		Id:              item.ID,
		OrderId:         item.OrderID,
		ProductId:       item.ProductID,
		ProductName:     item.ProductName,
		ProductSku:      item.ProductSKU,
		Quantity:        int32(item.Quantity),
		UnitPriceAmount: int32(item.UnitPriceAmount),
		LineTotalAmount: int32(item.LineTotalAmount),
		Currency:        item.Currency,
		CreatedAt:       item.CreatedAt.Format(time.RFC3339),
	}
}
