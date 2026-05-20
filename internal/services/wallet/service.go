package servicewallet

import (
	"context"
	"encoding/json"
	"strings"

	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	interfacewallet "3za-digital/internal/interfaces/wallet"
	"3za-digital/pkg/filter"
	"3za-digital/utils"

	"gorm.io/gorm"
)

type WalletService struct {
	Repo interfacewallet.RepoWalletInterface
}

func NewWalletService(repo interfacewallet.RepoWalletInterface) *WalletService {
	return &WalletService{Repo: repo}
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
	amount, err := normalizePositiveAmount(req.Amount)
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
		Method:    domainwallet.DepositMethodPaymentGateway,
		Provider:  strings.TrimSpace(req.Provider),
		Metadata:  json.RawMessage(`{}`),
		CreatedBy: &userID,
	}
	return s.Repo.CreatePaymentGatewayDeposit(ctx, deposit)
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

func (s *WalletService) AdminTopup(ctx context.Context, userID string, req dto.AdminWalletTopupRequest, actorID string) (domainwallet.DepositRequest, error) {
	amount, err := normalizePositiveAmount(req.Amount)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	userID = strings.TrimSpace(userID)
	actorID = strings.TrimSpace(actorID)
	if userID == "" || actorID == "" {
		return domainwallet.DepositRequest{}, domainwallet.ErrInvalidAmount
	}

	deposit := domainwallet.DepositRequest{
		Id:        utils.CreateUUID(),
		UserID:    userID,
		Amount:    amount,
		Status:    domainwallet.DepositStatusPaid,
		Method:    domainwallet.DepositMethodManualAdmin,
		Metadata:  json.RawMessage(`{}`),
		CreatedBy: &actorID,
	}
	return s.Repo.CreateManualTopup(ctx, deposit, strings.TrimSpace(req.Description))
}

func (s *WalletService) AdminAdjust(ctx context.Context, userID string, req dto.AdminWalletAdjustRequest, actorID string) (domainwallet.WalletTransaction, error) {
	amount, err := normalizePositiveAmount(req.Amount)
	if err != nil {
		return domainwallet.WalletTransaction{}, err
	}
	direction := strings.ToLower(strings.TrimSpace(req.Direction))
	if direction != domainwallet.DirectionCredit && direction != domainwallet.DirectionDebit {
		return domainwallet.WalletTransaction{}, domainwallet.ErrInvalidDirection
	}
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(actorID) == "" {
		return domainwallet.WalletTransaction{}, domainwallet.ErrInvalidAmount
	}
	return s.Repo.AdjustWallet(ctx, userID, direction, amount, strings.TrimSpace(req.Description), strings.TrimSpace(actorID))
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
