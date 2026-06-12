package boqris

import (
	"context"
	"strings"

	domainwallet "3za-digital/internal/domain/wallet"
	interfacewallet "3za-digital/internal/interfaces/wallet"
)

type Provider struct {
	client *Client
}

func NewProvider(config Config) (*Provider, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Provider{client: client}, nil
}

func (p *Provider) Provider() string {
	return domainwallet.DepositProviderBOQRIS
}

func (p *Provider) GenerateQRIS(ctx context.Context, req interfacewallet.QRISGenerateRequest) (*interfacewallet.QRISGenerateResponse, error) {
	response, err := p.client.CreateTransaction(ctx, CreateTransactionRequest{
		Amount:    req.Amount,
		InvoiceNo: req.InvoiceNo,
	})
	if err != nil {
		return nil, err
	}

	return &interfacewallet.QRISGenerateResponse{
		Provider:         domainwallet.DepositProviderBOQRIS,
		TransactionID:    response.TransactionID,
		PaymentReference: response.TransactionID,
		Status:           response.Status,
		QRISString:       response.QRISDynamic,
		QRISImageURL:     response.QRURL,
		PayableAmount:    response.Amount.Int64(),
		ExpiresAt:        response.ExpiresAt,
		Raw:              response.Raw,
	}, nil
}

func (p *Provider) GetPaymentStatus(ctx context.Context, paymentReference string) (*interfacewallet.QRISPaymentStatusResponse, error) {
	response, err := p.client.GetTransaction(ctx, paymentReference)
	if err != nil {
		return nil, err
	}

	return &interfacewallet.QRISPaymentStatusResponse{
		Provider:         domainwallet.DepositProviderBOQRIS,
		RequestID:        response.TransactionID,
		PaymentReference: strings.TrimSpace(paymentReference),
		Status:           response.Status,
		Amount:           response.Amount.Int64(),
		PaidAt:           paidAtValue(response.PaidAt),
		Raw:              response.Raw,
	}, nil
}

func paidAtValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
