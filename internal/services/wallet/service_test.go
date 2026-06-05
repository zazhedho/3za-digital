package servicewallet

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"3za-digital/internal/authscope"
	domainprovider "3za-digital/internal/domain/provider"
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/qrisly"
	"3za-digital/pkg/filter"
	"3za-digital/utils"

	"gorm.io/gorm"
)

type walletRepoStub struct {
	deposit              domainwallet.DepositRequest
	deposits             []domainwallet.DepositRequest
	depositsUserID       string
	createdDeposit       domainwallet.DepositRequest
	updatedProvider      string
	updatedReference     string
	updatedStatus        string
	completedProvider    string
	completedReference   string
	completedAmount      string
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
	r.depositsUserID = userID
	return r.deposits, int64(len(r.deposits)), nil
}

func (r *walletRepoStub) GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	if r.deposit.Id == id {
		return r.deposit, nil
	}
	return domainwallet.DepositRequest{}, errors.New("not found")
}

func (r *walletRepoStub) GetDepositWithUserByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	return r.GetDepositByID(ctx, id)
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
	r.updatedProvider = provider
	r.updatedReference = paymentReference
	r.updatedStatus = status
	for _, deposit := range r.deposits {
		if deposit.Provider == provider && deposit.PaymentReference == paymentReference {
			deposit.Status = status
			return deposit, nil
		}
	}
	r.deposit.Status = status
	return r.deposit, nil
}

func (r *walletRepoStub) CompleteDepositByPaymentReference(ctx context.Context, provider, paymentReference, amount string, log domainwallet.PaymentGatewayLog) (domainwallet.DepositRequest, error) {
	r.completedProvider = provider
	r.completedReference = paymentReference
	r.completedAmount = amount
	for _, deposit := range r.deposits {
		if deposit.Provider == provider && deposit.PaymentReference == paymentReference {
			deposit.Status = domainwallet.DepositStatusPaid
			return deposit, nil
		}
	}
	r.deposit.Status = domainwallet.DepositStatusPaid
	return r.deposit, nil
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

type qrisProviderStub struct {
	req             qrisly.GenerateQRISRequest
	statusHistoryID string
	response        *qrisly.GenerateQRISResponse
	statusResponse  *qrisly.PaymentStatusResponse
	err             error
}

func (p *qrisProviderStub) GenerateQRIS(ctx context.Context, req qrisly.GenerateQRISRequest) (*qrisly.GenerateQRISResponse, error) {
	p.req = req
	if p.err != nil {
		return nil, p.err
	}
	return p.response, nil
}

func (p *qrisProviderStub) GetPaymentStatus(ctx context.Context, historyID string) (*qrisly.PaymentStatusResponse, error) {
	p.statusHistoryID = historyID
	if p.err != nil {
		return nil, p.err
	}
	return p.statusResponse, nil
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

func TestCreateDepositRejectsBelowMinimumAmount(t *testing.T) {
	repo := &walletRepoStub{}
	service := NewWalletService(repo)

	_, err := service.CreateDeposit(context.Background(), "user-1", dto.CreateDepositRequest{Amount: "9999"})
	if !errors.Is(err, domainwallet.ErrDepositBelowMinimum) {
		t.Fatalf("expected ErrDepositBelowMinimum, got %v", err)
	}
	if repo.createdDeposit.Id != "" {
		t.Fatalf("expected no created deposit, got %+v", repo.createdDeposit)
	}
}

func TestCreateDepositQRISAddsFeeUniqueCodeAndPayableMetadata(t *testing.T) {
	repo := &walletRepoStub{}
	qrisProvider := &qrisProviderStub{response: &qrisly.GenerateQRISResponse{
		Data: qrisly.GenerateQRISData{
			HistoryID:      utils.FlexibleString("1778"),
			QRISImageURL:   "https://qris.test/image.png",
			OriginalAmount: utils.FlexibleInt64(105000),
			FinalAmount:    utils.FlexibleInt64(105003),
			PaymentStatus:  "unpaid",
			ExpiryTime:     "2026-03-03 14:52:52",
			MerchantName:   "ABC Store",
		},
	}}
	service := NewWalletService(repo).WithQRISProvider(qrisProvider)

	_, err := service.CreateDeposit(context.Background(), "user-1", dto.CreateDepositRequest{
		Amount:   "100000",
		Provider: "qris",
	})
	if err != nil {
		t.Fatalf("CreateDeposit returned error: %v", err)
	}
	if repo.createdDeposit.Amount != "100000.00" {
		t.Fatalf("expected credit amount 100000.00, got %q", repo.createdDeposit.Amount)
	}
	if repo.createdDeposit.Method != domainwallet.DepositMethodPaymentGateway {
		t.Fatalf("expected payment_gateway method, got %q", repo.createdDeposit.Method)
	}
	if repo.createdDeposit.Provider != domainwallet.DepositProviderQRISLY {
		t.Fatalf("expected qrisly provider, got %q", repo.createdDeposit.Provider)
	}
	if repo.createdDeposit.PaymentReference != "1778" {
		t.Fatalf("expected history id payment reference, got %q", repo.createdDeposit.PaymentReference)
	}
	if qrisProvider.req.Amount != 105000 {
		t.Fatalf("expected QRISLY amount 105000, got %d", qrisProvider.req.Amount)
	}
	if !qrisProvider.req.UniqueAmount {
		t.Fatal("expected unique amount request")
	}

	var metadata map[string]string
	if err := json.Unmarshal(repo.createdDeposit.Metadata, &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if metadata["fee_percent"] != "5" {
		t.Fatalf("expected fee percent 5, got %q", metadata["fee_percent"])
	}
	if metadata["fee_amount"] != "5000.00" {
		t.Fatalf("expected fee amount 5000.00, got %q", metadata["fee_amount"])
	}
	if metadata["unique_code"] != "3" || metadata["unique_code_amount"] != "3.00" {
		t.Fatalf("expected unique amount 3.00, got %+v", metadata)
	}
	if metadata["payable_amount"] != "105003.00" {
		t.Fatalf("expected payable amount 105003.00, got %q", metadata["payable_amount"])
	}
	if metadata["qrisly_history_id"] != "1778" {
		t.Fatalf("expected history metadata, got %+v", metadata)
	}
}

func TestGetDepositsUsesEmptyUserFilterForAdminList(t *testing.T) {
	repo := &walletRepoStub{}
	service := NewWalletService(repo)

	_, _, err := service.GetDeposits(context.Background(), filter.BaseParams{})
	if err != nil {
		t.Fatalf("GetDeposits returned error: %v", err)
	}
	if repo.depositsUserID != "" {
		t.Fatalf("expected empty user filter for admin list, got %q", repo.depositsUserID)
	}
}

func TestGetMyDepositsSyncsQRISLYUnpaidStatusForDisplay(t *testing.T) {
	repo := &walletRepoStub{deposits: []domainwallet.DepositRequest{{
		Id:               "deposit-1",
		UserID:           "member-1",
		Amount:           "10000.00",
		Status:           domainwallet.DepositStatusPending,
		Provider:         domainwallet.DepositProviderQRISLY,
		PaymentReference: "2847",
		Metadata:         utilsJSON(map[string]string{"payable_amount": "10002.00", "qrisly_history_id": "2847"}),
	}}}
	qrisProvider := &qrisProviderStub{statusResponse: &qrisly.PaymentStatusResponse{
		Data: qrisly.PaymentStatusData{
			HistoryID:     utils.FlexibleString("2847"),
			PaymentStatus: "unpaid",
			Amount:        utils.FlexibleInt64(10002),
		},
	}}
	service := NewWalletService(repo).WithQRISProvider(qrisProvider)

	deposits, _, err := service.GetMyDeposits(testUserContext("member-1"), filter.BaseParams{})
	if err != nil {
		t.Fatalf("GetMyDeposits returned error: %v", err)
	}
	if qrisProvider.statusHistoryID != "2847" {
		t.Fatalf("expected QRISLY history id 2847, got %q", qrisProvider.statusHistoryID)
	}
	if deposits[0].Status != domainwallet.DepositStatusPending {
		t.Fatalf("expected local pending status, got %q", deposits[0].Status)
	}
	var metadata map[string]string
	if err := json.Unmarshal(deposits[0].Metadata, &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if metadata["qrisly_status"] != "unpaid" {
		t.Fatalf("expected qrisly_status unpaid, got %+v", metadata)
	}
	if repo.updatedStatus != "" || repo.completedAmount != "" {
		t.Fatalf("expected no final repo update, got status=%q amount=%q", repo.updatedStatus, repo.completedAmount)
	}
}

func TestGetMyDepositsCompletesPaidQRISLYDeposit(t *testing.T) {
	repo := &walletRepoStub{deposits: []domainwallet.DepositRequest{{
		Id:               "deposit-1",
		UserID:           "member-1",
		Amount:           "10000.00",
		Status:           domainwallet.DepositStatusPending,
		Provider:         domainwallet.DepositProviderQRISLY,
		PaymentReference: "2847",
		Metadata:         utilsJSON(map[string]string{"payable_amount": "10002.00", "qrisly_history_id": "2847"}),
	}}}
	qrisProvider := &qrisProviderStub{statusResponse: &qrisly.PaymentStatusResponse{
		Data: qrisly.PaymentStatusData{
			HistoryID:     utils.FlexibleString("2847"),
			PaymentStatus: "paid",
			Amount:        utils.FlexibleInt64(10002),
		},
	}}
	service := NewWalletService(repo).WithQRISProvider(qrisProvider)

	deposits, _, err := service.GetMyDeposits(testUserContext("member-1"), filter.BaseParams{})
	if err != nil {
		t.Fatalf("GetMyDeposits returned error: %v", err)
	}
	if deposits[0].Status != domainwallet.DepositStatusPaid {
		t.Fatalf("expected paid status, got %q", deposits[0].Status)
	}
	if repo.completedProvider != domainwallet.DepositProviderQRISLY || repo.completedReference != "2847" || repo.completedAmount != "10002.00" {
		t.Fatalf("expected completion by qrisly ref amount, got provider=%q ref=%q amount=%q", repo.completedProvider, repo.completedReference, repo.completedAmount)
	}
}

func TestGetMyDepositsExpiresQRISLYDeposit(t *testing.T) {
	repo := &walletRepoStub{deposits: []domainwallet.DepositRequest{{
		Id:               "deposit-1",
		UserID:           "member-1",
		Amount:           "10000.00",
		Status:           domainwallet.DepositStatusPending,
		Provider:         domainwallet.DepositProviderQRISLY,
		PaymentReference: "2847",
		Metadata:         utilsJSON(map[string]string{"payable_amount": "10002.00", "qrisly_history_id": "2847"}),
	}}}
	qrisProvider := &qrisProviderStub{statusResponse: &qrisly.PaymentStatusResponse{
		Data: qrisly.PaymentStatusData{
			HistoryID:     utils.FlexibleString("2847"),
			PaymentStatus: "expired",
			Amount:        utils.FlexibleInt64(10002),
		},
	}}
	service := NewWalletService(repo).WithQRISProvider(qrisProvider)

	deposits, _, err := service.GetMyDeposits(testUserContext("member-1"), filter.BaseParams{})
	if err != nil {
		t.Fatalf("GetMyDeposits returned error: %v", err)
	}
	if deposits[0].Status != domainwallet.DepositStatusExpired {
		t.Fatalf("expected expired status, got %q", deposits[0].Status)
	}
	if repo.updatedProvider != domainwallet.DepositProviderQRISLY || repo.updatedReference != "2847" || repo.updatedStatus != domainwallet.DepositStatusExpired {
		t.Fatalf("expected expired update, got provider=%q ref=%q status=%q", repo.updatedProvider, repo.updatedReference, repo.updatedStatus)
	}
}

func TestGetDepositByIDUsesAdminDetailRepo(t *testing.T) {
	repo := &walletRepoStub{deposit: domainwallet.DepositRequest{
		Id:     "deposit-1",
		UserID: "member-1",
		Amount: "10000.00",
		Status: domainwallet.DepositStatusPending,
	}}
	service := NewWalletService(repo)

	deposit, err := service.GetDepositByID(context.Background(), "deposit-1")
	if err != nil {
		t.Fatalf("GetDepositByID returned error: %v", err)
	}
	if deposit.Id != "deposit-1" {
		t.Fatalf("expected deposit-1, got %q", deposit.Id)
	}
}

func TestGetMyDepositsUsesUserIDFromContext(t *testing.T) {
	repo := &walletRepoStub{}
	service := NewWalletService(repo)

	_, _, err := service.GetMyDeposits(testUserContext("member-1"), filter.BaseParams{})
	if err != nil {
		t.Fatalf("GetMyDeposits returned error: %v", err)
	}
	if repo.depositsUserID != "member-1" {
		t.Fatalf("expected user filter from context, got %q", repo.depositsUserID)
	}
}

func TestGetMyDepositsRejectsMissingUserContext(t *testing.T) {
	repo := &walletRepoStub{}
	service := NewWalletService(repo)

	_, _, err := service.GetMyDeposits(context.Background(), filter.BaseParams{})
	if !errors.Is(err, domainwallet.ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestGetMyDepositByIDUsesUserIDFromContext(t *testing.T) {
	repo := &walletRepoStub{deposit: domainwallet.DepositRequest{
		Id:     "deposit-1",
		UserID: "member-1",
		Amount: "10000.00",
		Status: domainwallet.DepositStatusPending,
	}}
	service := NewWalletService(repo)

	deposit, err := service.GetMyDepositByID(testUserContext("member-1"), "deposit-1")
	if err != nil {
		t.Fatalf("GetMyDepositByID returned error: %v", err)
	}
	if deposit.Id != "deposit-1" {
		t.Fatalf("expected deposit-1, got %q", deposit.Id)
	}
}

func TestGetMyDepositByIDRejectsOtherUser(t *testing.T) {
	repo := &walletRepoStub{deposit: domainwallet.DepositRequest{
		Id:     "deposit-1",
		UserID: "member-1",
		Amount: "10000.00",
		Status: domainwallet.DepositStatusPending,
	}}
	service := NewWalletService(repo)

	_, err := service.GetMyDepositByID(testUserContext("member-2"), "deposit-1")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestGetMyDepositByIDRejectsMissingUserContext(t *testing.T) {
	repo := &walletRepoStub{}
	service := NewWalletService(repo)

	_, err := service.GetMyDepositByID(context.Background(), "deposit-1")
	if !errors.Is(err, domainwallet.ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestHandlePaymentWebhookDisabledByDefault(t *testing.T) {
	service := NewWalletService(&walletRepoStub{})

	_, err := service.HandlePaymentWebhook(context.Background(), "h2h", dto.PaymentWebhookRequest{
		PaymentReference: "ref-1",
		Amount:           "10000",
		Status:           "paid",
	})
	if !errors.Is(err, domainwallet.ErrPaymentWebhookDisabled) {
		t.Fatalf("expected ErrPaymentWebhookDisabled, got %v", err)
	}
}

func TestHandlePaymentWebhookRequiresSecretWhenEnabled(t *testing.T) {
	t.Setenv("PAYMENT_WEBHOOK_ENABLED", "true")
	service := NewWalletService(&walletRepoStub{})

	_, err := service.HandlePaymentWebhook(context.Background(), "midtrans", dto.PaymentWebhookRequest{
		PaymentReference: "ref-1",
		Amount:           "10000",
		Status:           "paid",
		Signature:        "signature",
	})
	if !errors.Is(err, domainwallet.ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestAdminTopupRequiresMainBalanceProvider(t *testing.T) {
	service := NewWalletService(&walletRepoStub{})

	_, err := service.AdminTopup(testActorContext("admin-1"), "user-1", dto.AdminWalletTopupRequest{Amount: "10000"})
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

	_, err := service.AdminTopup(testActorContext("admin-1"), "user-1", dto.AdminWalletTopupRequest{
		DepositRequestID: "deposit-1",
		Amount:           "10000.00",
	})
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

	_, err := service.AdminApproveDeposit(testActorContext("admin-1"), "deposit-1", dto.AdminDepositApproveRequest{
		Amount:      "150000.00",
		Description: "paid by transfer",
	})
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

	_, err := service.AdminTopup(testActorContext("admin-1"), "user-1", dto.AdminWalletTopupRequest{
		DepositRequestID: "deposit-1",
		Amount:           "9000",
	})
	if !errors.Is(err, domainwallet.ErrDepositAmountMismatch) {
		t.Fatalf("expected ErrDepositAmountMismatch, got %v", err)
	}
}

func TestAdminTopupAllowsZeroMainBalanceToReachInsufficientMainBalanceCheck(t *testing.T) {
	repo := &walletRepoStub{createManualTopupErr: domainwallet.ErrInsufficientMainBalance}
	service := NewWalletService(repo, mainBalanceProviderStub{balance: "0"})

	_, err := service.AdminTopup(testActorContext("admin-1"), "user-1", dto.AdminWalletTopupRequest{Amount: "10000"})
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

	_, err := service.AdminAdjust(testActorContext("admin-1"), "user-1", dto.AdminWalletAdjustRequest{
		Amount:      "5000",
		Direction:   domainwallet.DirectionCredit,
		Description: "manual correction",
	})
	if err != nil {
		t.Fatalf("AdminAdjust returned error: %v", err)
	}
	if repo.adjustLimit != "75000.00" {
		t.Fatalf("expected H2H balance limit 75000.00, got %q", repo.adjustLimit)
	}
}

func testActorContext(userID string) context.Context {
	return authscope.WithContext(context.Background(), authscope.New(userID, "admin", "admin", nil))
}

func testUserContext(userID string) context.Context {
	return authscope.WithContext(context.Background(), authscope.New(userID, "member", "member", nil))
}

func utilsJSON(value interface{}) json.RawMessage {
	raw, _ := json.Marshal(value)
	return raw
}
