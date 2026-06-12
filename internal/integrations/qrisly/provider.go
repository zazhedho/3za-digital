package qrisly

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
	return domainwallet.DepositProviderQRISLY
}

func (p *Provider) GenerateQRIS(ctx context.Context, req interfacewallet.QRISGenerateRequest) (*interfacewallet.QRISGenerateResponse, error) {
	response, err := p.client.GenerateQRIS(ctx, GenerateQRISRequest{
		Amount:       req.Amount,
		UniqueAmount: true,
	})
	if err != nil {
		return nil, err
	}

	return &interfacewallet.QRISGenerateResponse{
		Provider:         domainwallet.DepositProviderQRISLY,
		TransactionID:    response.Data.HistoryID.String(),
		PaymentReference: response.Data.HistoryID.String(),
		Status:           qrislyStatus(response.Data.PaymentStatus, response.Data.PaymentStatus),
		MerchantName:     response.Data.MerchantValue(),
		QRISString:       response.Data.QRISString,
		QRISImageURL:     response.Data.ImageValue(),
		PayableAmount:    response.Data.PayableAmount(),
		ExpiresAt:        response.Data.ExpiryValue(),
		Raw:              response.Raw,
	}, nil
}

func (p *Provider) GetPaymentStatus(ctx context.Context, paymentReference string) (*interfacewallet.QRISPaymentStatusResponse, error) {
	response, err := p.client.GetPaymentStatus(ctx, paymentReference)
	if err != nil {
		return nil, err
	}

	return &interfacewallet.QRISPaymentStatusResponse{
		Provider:         domainwallet.DepositProviderQRISLY,
		RequestID:        response.Data.HistoryID.String(),
		PaymentReference: strings.TrimSpace(paymentReference),
		Status:           qrislyStatus(response.Data.PaymentStatus, response.Data.Status),
		Amount:           response.Data.Amount.Int64(),
		PaidAt:           response.Data.PaidAt.String(),
		Raw:              response.Raw,
	}, nil
}

func qrislyStatus(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
