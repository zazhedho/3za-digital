package interfacewallet

import (
	"context"

	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	"3za-digital/pkg/filter"
)

type ServiceWalletInterface interface {
	GetMyWallet(ctx context.Context, userID string) (domainwallet.Wallet, error)
	GetMyTransactions(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.WalletTransaction, int64, error)
	GetWallets(ctx context.Context, params filter.BaseParams) ([]domainwallet.Wallet, int64, error)
	CreateDeposit(ctx context.Context, userID string, req dto.CreateDepositRequest) (domainwallet.DepositRequest, error)
	GetMyDeposits(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error)
	GetMyDepositByID(ctx context.Context, userID, id string) (domainwallet.DepositRequest, error)
	AdminTopup(ctx context.Context, userID string, req dto.AdminWalletTopupRequest) (domainwallet.DepositRequest, error)
	AdminApproveDeposit(ctx context.Context, depositID string, req dto.AdminDepositApproveRequest) (domainwallet.DepositRequest, error)
	AdminAdjust(ctx context.Context, userID string, req dto.AdminWalletAdjustRequest) (domainwallet.WalletTransaction, error)
	HandlePaymentWebhook(ctx context.Context, provider string, req dto.PaymentWebhookRequest) (domainwallet.DepositRequest, error)
}
