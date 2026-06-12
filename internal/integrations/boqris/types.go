package boqris

import (
	"encoding/json"

	"3za-digital/utils"
)

type CreateTransactionRequest struct {
	MerchantID string `json:"merchant_id"`
	Amount     int64  `json:"amount"`
	InvoiceNo  string `json:"invoice_no"`
}

type CreateTransactionResponse struct {
	TransactionID string              `json:"transaction_id"`
	Status        string              `json:"status"`
	MerchantID    string              `json:"merchant_id"`
	BaseAmount    utils.FlexibleInt64 `json:"base_amount"`
	UniqueCode    utils.FlexibleInt64 `json:"unique_code"`
	Amount        utils.FlexibleInt64 `json:"amount"`
	InvoiceNo     string              `json:"invoice_no"`
	QRISDynamic   string              `json:"qris_dynamic"`
	QRURL         string              `json:"qr_url"`
	ExpiresAt     string              `json:"expires_at"`
	PaidAt        *string             `json:"paid_at"`
	CreatedAt     string              `json:"created_at"`
	Raw           json.RawMessage     `json:"-"`
}

type APIError struct {
	HTTPStatus int
	Message    string
	Raw        json.RawMessage
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "boqris api error"
}
