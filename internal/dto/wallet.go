package dto

type CreateDepositRequest struct {
	Amount   string `json:"amount" binding:"required"`
	Provider string `json:"provider"`
}

type AdminWalletTopupRequest struct {
	Amount      string `json:"amount" binding:"required"`
	Description string `json:"description"`
}

type AdminWalletAdjustRequest struct {
	Amount      string `json:"amount" binding:"required"`
	Direction   string `json:"direction" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type PaymentWebhookRequest struct {
	EventType        string                 `json:"event_type"`
	RequestID        string                 `json:"request_id"`
	PaymentReference string                 `json:"payment_reference" binding:"required"`
	Amount           string                 `json:"amount" binding:"required"`
	Status           string                 `json:"status" binding:"required"`
	Signature        string                 `json:"signature"`
	Payload          map[string]interface{} `json:"payload"`
}
