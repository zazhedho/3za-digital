package interfacewallet

import (
	"context"

	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/pkg/filter"
)

type RepoWalletInterface interface {
	GetWalletByUserID(ctx context.Context, userID string) (domainwallet.Wallet, error)
	EnsureWallet(ctx context.Context, userID string) (domainwallet.Wallet, error)
	GetTransactions(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.WalletTransaction, int64, error)
	GetWallets(ctx context.Context, params filter.BaseParams) ([]domainwallet.Wallet, int64, error)
	GetDeposits(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error)
	GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error)
	GetDepositWithUserByID(ctx context.Context, id string) (domainwallet.DepositRequest, error)
	CreateDepositRequest(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error)
	CreateManualTopup(ctx context.Context, deposit domainwallet.DepositRequest, description string, mainBalanceLimit string) (domainwallet.DepositRequest, error)
	ApproveManualTopup(ctx context.Context, deposit domainwallet.DepositRequest, description string, mainBalanceLimit string) (domainwallet.DepositRequest, error)
	AdjustWallet(ctx context.Context, userID, direction, amount, description, createdBy string, mainBalanceLimit string) (domainwallet.WalletTransaction, error)
	UpdateDepositStatusByPaymentReference(ctx context.Context, provider, paymentReference, status string, log domainwallet.PaymentGatewayLog) (domainwallet.DepositRequest, error)
	CompleteDepositByPaymentReference(ctx context.Context, provider, paymentReference, amount string, log domainwallet.PaymentGatewayLog) (domainwallet.DepositRequest, error)
}
