package interfacepaymentgateway

import "context"

type CreateInvoiceRequest struct {
	DepositID string
	UserID    string
	Amount    string
	Provider  string
	Metadata  map[string]interface{}
}

type CreateInvoiceResponse struct {
	PaymentReference string
	PaymentURL       string
	ExpiredAt        string
	Raw              map[string]interface{}
}

type WebhookVerificationRequest struct {
	Provider  string
	Signature string
	Payload   []byte
}

type Gateway interface {
	CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (CreateInvoiceResponse, error)
	VerifyWebhook(ctx context.Context, req WebhookVerificationRequest) error
}
