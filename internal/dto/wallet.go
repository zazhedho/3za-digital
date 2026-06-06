package dto

type CreateDepositRequest struct {
	Amount   string `json:"amount" binding:"required"`
	Provider string `json:"provider"`
}

type DepositSettingsResponse struct {
	MinimumAmount      string `json:"minimum_amount"`
	QRISFeePercent     string `json:"qris_fee_percent"`
	QRISStaticImageURL string `json:"qris_static_image_url"`
}

type AdminWalletTopupRequest struct {
	Amount           string `json:"amount"`
	DepositRequestID string `json:"deposit_request_id"`
	Description      string `json:"description"`
}

type AdminDepositApproveRequest struct {
	Amount      string `json:"amount"`
	Description string `json:"description"`
}

type AdminDepositCancelRequest struct {
	Reason string `json:"reason"`
}

type AdminDepositStatusRequest struct {
	Status      string `json:"status" binding:"required"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
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

type QRISLYWebhookRequest struct {
	Event     string            `json:"event"`
	Timestamp string            `json:"timestamp"`
	Data      QRISLYWebhookData `json:"data"`
}

type QRISLYWebhookData struct {
	HistoryID       interface{} `json:"history_id"`
	QRISID          interface{} `json:"qris_id"`
	Amount          interface{} `json:"amount"`
	OriginalAmount  interface{} `json:"original_amount"`
	Status          string      `json:"status"`
	PaidAt          string      `json:"paid_at"`
	ExpiredAt       string      `json:"expired_at"`
	PaymentMethod   string      `json:"payment_method"`
	PaymentProvider string      `json:"payment_provider"`
	CreatedAt       string      `json:"created_at"`
}
