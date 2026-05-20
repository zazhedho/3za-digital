package servicewallet

import (
	"context"
	"encoding/json"
	"strings"

	"3za-digital/internal/authscope"
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	interfacewallet "3za-digital/internal/interfaces/wallet"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/money"
	"3za-digital/utils"

	"gorm.io/gorm"
)

type WalletService struct {
	Repo                interfacewallet.RepoWalletInterface
	MainBalanceProvider interfacewallet.MainBalanceProvider
}

func NewWalletService(repo interfacewallet.RepoWalletInterface, mainBalanceProviders ...interfacewallet.MainBalanceProvider) *WalletService {
	service := &WalletService{Repo: repo}
	if len(mainBalanceProviders) > 0 {
		service.MainBalanceProvider = mainBalanceProviders[0]
	}
	return service
}

func (s *WalletService) GetMyWallet(ctx context.Context, userID string) (domainwallet.Wallet, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domainwallet.Wallet{}, domainwallet.ErrInvalidAmount
	}
	return s.Repo.EnsureWallet(ctx, userID)
}

func (s *WalletService) GetMyTransactions(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.WalletTransaction, int64, error) {
	return s.Repo.GetTransactions(ctx, strings.TrimSpace(userID), params)
}

func (s *WalletService) GetWallets(ctx context.Context, params filter.BaseParams) ([]domainwallet.Wallet, int64, error) {
	return s.Repo.GetWallets(ctx, params)
}

func (s *WalletService) CreateDeposit(ctx context.Context, userID string, req dto.CreateDepositRequest) (domainwallet.DepositRequest, error) {
	amount, err := money.NormalizePositive(req.Amount)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domainwallet.DepositRequest{}, domainwallet.ErrInvalidAmount
	}

	deposit := domainwallet.DepositRequest{
		Id:        utils.CreateUUID(),
		UserID:    userID,
		Amount:    amount,
		Status:    domainwallet.DepositStatusPending,
		Method:    domainwallet.DepositMethodManualAdmin,
		Provider:  strings.TrimSpace(req.Provider),
		Metadata:  utils.EmptyJSON(),
		CreatedBy: &userID,
	}
	return s.Repo.CreateDepositRequest(ctx, deposit)
}

func (s *WalletService) GetMyDeposits(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	return s.Repo.GetDeposits(ctx, strings.TrimSpace(userID), params)
}

func (s *WalletService) GetMyDepositByID(ctx context.Context, userID, id string) (domainwallet.DepositRequest, error) {
	deposit, err := s.Repo.GetDepositByID(ctx, id)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	if deposit.UserID != strings.TrimSpace(userID) {
		return domainwallet.DepositRequest{}, gorm.ErrRecordNotFound
	}
	return deposit, nil
}

func (s *WalletService) AdminTopup(ctx context.Context, userID string, req dto.AdminWalletTopupRequest) (domainwallet.DepositRequest, error) {
	userID = strings.TrimSpace(userID)
	actorID := authscope.FromContext(ctx).ActorUserID()
	if userID == "" || actorID == "" {
		return domainwallet.DepositRequest{}, domainwallet.ErrInvalidAmount
	}

	depositID := strings.TrimSpace(req.DepositRequestID)
	deposit := domainwallet.DepositRequest{
		Id:        utils.CreateUUID(),
		UserID:    userID,
		Status:    domainwallet.DepositStatusPaid,
		Method:    domainwallet.DepositMethodManualAdmin,
		Metadata:  utils.EmptyJSON(),
		CreatedBy: &actorID,
	}
	if depositID != "" {
		existingDeposit, err := s.Repo.GetDepositByID(ctx, depositID)
		if err != nil {
			return domainwallet.DepositRequest{}, err
		}
		if strings.TrimSpace(existingDeposit.UserID) != userID {
			return domainwallet.DepositRequest{}, gorm.ErrRecordNotFound
		}
		if existingDeposit.Status != domainwallet.DepositStatusPending {
			return domainwallet.DepositRequest{}, domainwallet.ErrDepositAlreadyFinal
		}
		if strings.TrimSpace(req.Amount) != "" && money.NormalizeOrTrim(req.Amount) != money.NormalizeOrTrim(existingDeposit.Amount) {
			return domainwallet.DepositRequest{}, domainwallet.ErrDepositAmountMismatch
		}
		deposit = existingDeposit
		deposit.Amount = money.NormalizeOrTrim(existingDeposit.Amount)
		deposit.Method = domainwallet.DepositMethodManualAdmin
		deposit.CreatedBy = &actorID
	} else {
		normalizedAmount, err := money.NormalizePositive(req.Amount)
		if err != nil {
			return domainwallet.DepositRequest{}, err
		}
		deposit.Amount = normalizedAmount
	}

	mainBalance, err := s.currentMainBalance(ctx)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	if depositID != "" {
		return s.Repo.ApproveManualTopup(ctx, deposit, strings.TrimSpace(req.Description), mainBalance)
	}
	return s.Repo.CreateManualTopup(ctx, deposit, strings.TrimSpace(req.Description), mainBalance)
}

func (s *WalletService) AdminApproveDeposit(ctx context.Context, depositID string, req dto.AdminDepositApproveRequest) (domainwallet.DepositRequest, error) {
	actorID := authscope.FromContext(ctx).ActorUserID()
	if strings.TrimSpace(depositID) == "" || actorID == "" {
		return domainwallet.DepositRequest{}, domainwallet.ErrInvalidAmount
	}

	deposit, err := s.Repo.GetDepositByID(ctx, strings.TrimSpace(depositID))
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	if deposit.Status != domainwallet.DepositStatusPending {
		return domainwallet.DepositRequest{}, domainwallet.ErrDepositAlreadyFinal
	}
	if strings.TrimSpace(req.Amount) != "" && money.NormalizeOrTrim(req.Amount) != money.NormalizeOrTrim(deposit.Amount) {
		return domainwallet.DepositRequest{}, domainwallet.ErrDepositAmountMismatch
	}
	deposit.Amount = money.NormalizeOrTrim(deposit.Amount)
	deposit.Method = domainwallet.DepositMethodManualAdmin
	deposit.CreatedBy = &actorID

	mainBalance, err := s.currentMainBalance(ctx)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	return s.Repo.ApproveManualTopup(ctx, deposit, strings.TrimSpace(req.Description), mainBalance)
}

func (s *WalletService) AdminAdjust(ctx context.Context, userID string, req dto.AdminWalletAdjustRequest) (domainwallet.WalletTransaction, error) {
	amount, err := money.NormalizePositive(req.Amount)
	if err != nil {
		return domainwallet.WalletTransaction{}, err
	}
	direction := utils.NormalizeKey(req.Direction)
	if direction != domainwallet.DirectionCredit && direction != domainwallet.DirectionDebit {
		return domainwallet.WalletTransaction{}, domainwallet.ErrInvalidDirection
	}
	actorID := authscope.FromContext(ctx).ActorUserID()
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(actorID) == "" {
		return domainwallet.WalletTransaction{}, domainwallet.ErrInvalidAmount
	}
	mainBalance := ""
	if direction == domainwallet.DirectionCredit {
		var err error
		mainBalance, err = s.currentMainBalance(ctx)
		if err != nil {
			return domainwallet.WalletTransaction{}, err
		}
	}
	return s.Repo.AdjustWallet(ctx, userID, direction, amount, strings.TrimSpace(req.Description), strings.TrimSpace(actorID), mainBalance)
}

func (s *WalletService) HandlePaymentWebhook(ctx context.Context, provider string, req dto.PaymentWebhookRequest) (domainwallet.DepositRequest, error) {
	provider = strings.TrimSpace(provider)
	if err := verifyWebhookSignature(provider, req); err != nil {
		return domainwallet.DepositRequest{}, err
	}

	status := normalizeGatewayStatus(req.Status)
	payload := req.Payload
	if payload == nil {
		payload = map[string]interface{}{}
	}
	rawPayload, _ := json.Marshal(payload)
	log := domainwallet.PaymentGatewayLog{
		Id:               utils.CreateUUID(),
		Provider:         provider,
		EventType:        strings.TrimSpace(req.EventType),
		RequestID:        strings.TrimSpace(req.RequestID),
		PaymentReference: strings.TrimSpace(req.PaymentReference),
		Signature:        strings.TrimSpace(req.Signature),
		Status:           strings.TrimSpace(req.Status),
		Payload:          rawPayload,
	}
	if !isDepositWebhookStatus(status) {
		log.ErrorMessage = "unknown payment status"
		return domainwallet.DepositRequest{}, domainwallet.ErrInvalidDirection
	}
	if status != domainwallet.DepositStatusPaid {
		return s.Repo.UpdateDepositStatusByPaymentReference(ctx, provider, strings.TrimSpace(req.PaymentReference), status, log)
	}
	return s.Repo.CompleteDepositByPaymentReference(ctx, provider, strings.TrimSpace(req.PaymentReference), req.Amount, log)
}

var _ interfacewallet.ServiceWalletInterface = (*WalletService)(nil)

func (s *WalletService) currentMainBalance(ctx context.Context) (string, error) {
	if s.MainBalanceProvider == nil {
		return "", domainwallet.ErrMainBalanceUnavailable
	}
	snapshot, err := s.MainBalanceProvider.GetH2HBalance(ctx)
	if err != nil {
		return "", err
	}
	balance, err := money.Normalize(snapshot.Balance)
	if err != nil {
		return "", domainwallet.ErrMainBalanceUnavailable
	}
	cents, err := money.ParseCents(balance)
	if err != nil || cents < 0 {
		return "", domainwallet.ErrMainBalanceUnavailable
	}
	return balance, nil
}
