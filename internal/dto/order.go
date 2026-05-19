package dto

type CreateOrderRequest struct {
	ServiceID string `json:"service_id" binding:"required"`
	Target    string `json:"target" binding:"required"`
	Quantity  int64  `json:"quantity" binding:"required,gte=1"`
}
