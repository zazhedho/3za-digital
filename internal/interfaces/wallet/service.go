package interfacewallet

import (
	"context"
	"encoding/json"

	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	"3za-digital/pkg/filter"
)

type QRISGenerateRequest struct {
	Amount    int64
	InvoiceNo string
}

type QRISGenerateResponse struct {
	Provider         string
	TransactionID    string
	PaymentReference string
	Status           string
	MerchantName     string
	QRISString       string
	QRISImageURL     string
	PayableAmount    int64
	ExpiresAt        string
	Raw              json.RawMessage
}

type QRISPaymentStatusResponse struct {
	Provider         string
	RequestID        string
	PaymentReference string
	Status           string
	Amount           int64
	PaidAt           string
	Raw              json.RawMessage
}

type QRISPaymentProvider interface {
	Provider() string
	GenerateQRIS(ctx context.Context, req QRISGenerateRequest) (*QRISGenerateResponse, error)
	GetPaymentStatus(ctx context.Context, paymentReference string) (*QRISPaymentStatusResponse, error)
}

type ServiceWalletInterface interface {
	GetMyWallet(ctx context.Context, userID string) (domainwallet.Wallet, error)
	GetMyTransactions(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.WalletTransaction, int64, error)
	GetWallets(ctx context.Context, params filter.BaseParams) ([]domainwallet.Wallet, int64, error)
	GetDepositSettings(ctx context.Context) (dto.DepositSettingsResponse, error)
	CreateDeposit(ctx context.Context, userID string, req dto.CreateDepositRequest) (domainwallet.DepositRequest, error)
	GetMyDeposits(ctx context.Context, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error)
	GetDeposits(ctx context.Context, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error)
	GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error)
	GetMyDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error)
	AdminTopup(ctx context.Context, userID string, req dto.AdminWalletTopupRequest) (domainwallet.DepositRequest, error)
	AdminApproveDeposit(ctx context.Context, depositID string, req dto.AdminDepositApproveRequest) (domainwallet.DepositRequest, error)
	AdminCancelDeposit(ctx context.Context, depositID string, req dto.AdminDepositCancelRequest) (domainwallet.DepositRequest, error)
	AdminAdjust(ctx context.Context, userID string, req dto.AdminWalletAdjustRequest) (domainwallet.WalletTransaction, error)
	HandlePaymentWebhook(ctx context.Context, provider string, req dto.PaymentWebhookRequest) (domainwallet.DepositRequest, error)
}
