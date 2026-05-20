package servicewallet

import (
	"context"
	"errors"
	"testing"

	domainprovider "3za-digital/internal/domain/provider"
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	"3za-digital/pkg/filter"
)

type walletRepoStub struct {
	deposit              domainwallet.DepositRequest
	createdDeposit       domainwallet.DepositRequest
	manualTopupLimit     string
	adjustLimit          string
	createManualTopupErr error
}

func (r *walletRepoStub) GetWalletByUserID(ctx context.Context, userID string) (domainwallet.Wallet, error) {
	return domainwallet.Wallet{}, nil
}

func (r *walletRepoStub) EnsureWallet(ctx context.Context, userID string) (domainwallet.Wallet, error) {
	return domainwallet.Wallet{}, nil
}

func (r *walletRepoStub) GetTransactions(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.WalletTransaction, int64, error) {
	return nil, 0, nil
}

func (r *walletRepoStub) GetWallets(ctx context.Context, params filter.BaseParams) ([]domainwallet.Wallet, int64, error) {
	return nil, 0, nil
}

func (r *walletRepoStub) GetDeposits(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	return nil, 0, nil
}

func (r *walletRepoStub) GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	if r.deposit.Id == id {
		return r.deposit, nil
	}
	return domainwallet.DepositRequest{}, errors.New("not found")
}

func (r *walletRepoStub) CreateDepositRequest(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	r.createdDeposit = deposit
	return deposit, nil
}

func (r *walletRepoStub) CreateManualTopup(ctx context.Context, deposit domainwallet.DepositRequest, description string, mainBalanceLimit string) (domainwallet.DepositRequest, error) {
	r.createdDeposit = deposit
	r.manualTopupLimit = mainBalanceLimit
	return deposit, r.createManualTopupErr
}

func (r *walletRepoStub) ApproveManualTopup(ctx context.Context, deposit domainwallet.DepositRequest, description string, mainBalanceLimit string) (domainwallet.DepositRequest, error) {
	r.createdDeposit = deposit
	r.manualTopupLimit = mainBalanceLimit
	return deposit, r.createManualTopupErr
}

func (r *walletRepoStub) AdjustWallet(ctx context.Context, userID, direction, amount, description, createdBy string, mainBalanceLimit string) (domainwallet.WalletTransaction, error) {
	r.adjustLimit = mainBalanceLimit
	return domainwallet.WalletTransaction{}, nil
}

func (r *walletRepoStub) UpdateDepositStatusByPaymentReference(ctx context.Context, provider, paymentReference, status string, log domainwallet.PaymentGatewayLog) (domainwallet.DepositRequest, error) {
	return domainwallet.DepositRequest{}, nil
}

func (r *walletRepoStub) CompleteDepositByPaymentReference(ctx context.Context, provider, paymentReference, amount string, log domainwallet.PaymentGatewayLog) (domainwallet.DepositRequest, error) {
	return domainwallet.DepositRequest{}, nil
}

type mainBalanceProviderStub struct {
	balance string
	err     error
}

func (p mainBalanceProviderStub) GetH2HBalance(ctx context.Context) (domainprovider.BalanceSnapshot, error) {
	if p.err != nil {
		return domainprovider.BalanceSnapshot{}, p.err
	}
	return domainprovider.BalanceSnapshot{Balance: p.balance}, nil
}

func TestCreateDepositUsesManualAdminPendingMethod(t *testing.T) {
	repo := &walletRepoStub{}
	service := NewWalletService(repo)

	_, err := service.CreateDeposit(context.Background(), "user-1", dto.CreateDepositRequest{Amount: "10000"})
	if err != nil {
		t.Fatalf("CreateDeposit returned error: %v", err)
	}
	if repo.createdDeposit.Status != domainwallet.DepositStatusPending {
		t.Fatalf("expected pending deposit, got %q", repo.createdDeposit.Status)
	}
	if repo.createdDeposit.Method != domainwallet.DepositMethodManualAdmin {
		t.Fatalf("expected manual_admin method, got %q", repo.createdDeposit.Method)
	}
}

func TestAdminTopupRequiresMainBalanceProvider(t *testing.T) {
	service := NewWalletService(&walletRepoStub{})

	_, err := service.AdminTopup(context.Background(), "user-1", dto.AdminWalletTopupRequest{Amount: "10000"}, "admin-1")
	if !errors.Is(err, domainwallet.ErrMainBalanceUnavailable) {
		t.Fatalf("expected ErrMainBalanceUnavailable, got %v", err)
	}
}

func TestAdminTopupApprovesPendingDepositWithMainBalanceLimit(t *testing.T) {
	repo := &walletRepoStub{deposit: domainwallet.DepositRequest{
		Id:     "deposit-1",
		UserID: "user-1",
		Amount: "10000",
		Status: domainwallet.DepositStatusPending,
		Method: domainwallet.DepositMethodManualAdmin,
	}}
	service := NewWalletService(repo, mainBalanceProviderStub{balance: "50000"})

	_, err := service.AdminTopup(context.Background(), "user-1", dto.AdminWalletTopupRequest{
		DepositRequestID: "deposit-1",
		Amount:           "10000.00",
	}, "admin-1")
	if err != nil {
		t.Fatalf("AdminTopup returned error: %v", err)
	}
	if repo.createdDeposit.Id != "deposit-1" {
		t.Fatalf("expected existing deposit to be approved, got %q", repo.createdDeposit.Id)
	}
	if repo.manualTopupLimit != "50000.00" {
		t.Fatalf("expected H2H balance limit 50000.00, got %q", repo.manualTopupLimit)
	}
}

func TestAdminApproveDepositUsesPendingDepositAndMainBalanceLimit(t *testing.T) {
	repo := &walletRepoStub{deposit: domainwallet.DepositRequest{
		Id:     "deposit-1",
		UserID: "user-1",
		Amount: "150000",
		Status: domainwallet.DepositStatusPending,
		Method: domainwallet.DepositMethodManualAdmin,
	}}
	service := NewWalletService(repo, mainBalanceProviderStub{balance: "200000"})

	_, err := service.AdminApproveDeposit(context.Background(), "deposit-1", dto.AdminDepositApproveRequest{
		Amount:      "150000.00",
		Description: "paid by transfer",
	}, "admin-1")
	if err != nil {
		t.Fatalf("AdminApproveDeposit returned error: %v", err)
	}
	if repo.createdDeposit.Id != "deposit-1" {
		t.Fatalf("expected existing deposit to be approved, got %q", repo.createdDeposit.Id)
	}
	if repo.manualTopupLimit != "200000.00" {
		t.Fatalf("expected H2H balance limit 200000.00, got %q", repo.manualTopupLimit)
	}
}

func TestAdminTopupRejectsDepositAmountMismatch(t *testing.T) {
	repo := &walletRepoStub{deposit: domainwallet.DepositRequest{
		Id:     "deposit-1",
		UserID: "user-1",
		Amount: "10000",
		Status: domainwallet.DepositStatusPending,
		Method: domainwallet.DepositMethodManualAdmin,
	}}
	service := NewWalletService(repo, mainBalanceProviderStub{balance: "50000"})

	_, err := service.AdminTopup(context.Background(), "user-1", dto.AdminWalletTopupRequest{
		DepositRequestID: "deposit-1",
		Amount:           "9000",
	}, "admin-1")
	if !errors.Is(err, domainwallet.ErrDepositAmountMismatch) {
		t.Fatalf("expected ErrDepositAmountMismatch, got %v", err)
	}
}

func TestAdminTopupAllowsZeroMainBalanceToReachInsufficientMainBalanceCheck(t *testing.T) {
	repo := &walletRepoStub{createManualTopupErr: domainwallet.ErrInsufficientMainBalance}
	service := NewWalletService(repo, mainBalanceProviderStub{balance: "0"})

	_, err := service.AdminTopup(context.Background(), "user-1", dto.AdminWalletTopupRequest{Amount: "10000"}, "admin-1")
	if !errors.Is(err, domainwallet.ErrInsufficientMainBalance) {
		t.Fatalf("expected ErrInsufficientMainBalance, got %v", err)
	}
	if repo.manualTopupLimit != "0.00" {
		t.Fatalf("expected H2H balance limit 0.00, got %q", repo.manualTopupLimit)
	}
}

func TestAdminCreditAdjustmentUsesMainBalanceLimit(t *testing.T) {
	repo := &walletRepoStub{}
	service := NewWalletService(repo, mainBalanceProviderStub{balance: "75000"})

	_, err := service.AdminAdjust(context.Background(), "user-1", dto.AdminWalletAdjustRequest{
		Amount:      "5000",
		Direction:   domainwallet.DirectionCredit,
		Description: "manual correction",
	}, "admin-1")
	if err != nil {
		t.Fatalf("AdminAdjust returned error: %v", err)
	}
	if repo.adjustLimit != "75000.00" {
		t.Fatalf("expected H2H balance limit 75000.00, got %q", repo.adjustLimit)
	}
}
