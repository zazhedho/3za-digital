package servicewallet

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"3za-digital/internal/authscope"
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	"3za-digital/internal/integrations/qrisly"
	interfaceappconfig "3za-digital/internal/interfaces/appconfig"
	interfacewallet "3za-digital/internal/interfaces/wallet"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/money"
	"3za-digital/utils"

	"gorm.io/gorm"
)

type WalletService struct {
	Repo                interfacewallet.RepoWalletInterface
	MainBalanceProvider interfacewallet.MainBalanceProvider
	ConfigService       interfaceappconfig.ServiceAppConfigInterface
	QRISProvider        QRISPaymentProvider
}

type QRISPaymentProvider interface {
	GenerateQRIS(ctx context.Context, req qrisly.GenerateQRISRequest) (*qrisly.GenerateQRISResponse, error)
}

func NewWalletService(repo interfacewallet.RepoWalletInterface, mainBalanceProviders ...interfacewallet.MainBalanceProvider) *WalletService {
	service := &WalletService{Repo: repo}
	if len(mainBalanceProviders) > 0 {
		service.MainBalanceProvider = mainBalanceProviders[0]
	}
	return service
}

func (s *WalletService) WithConfigService(configService interfaceappconfig.ServiceAppConfigInterface) *WalletService {
	s.ConfigService = configService
	return s
}

func (s *WalletService) WithQRISProvider(provider QRISPaymentProvider) *WalletService {
	s.QRISProvider = provider
	return s
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
	if belowMinimumDeposit(amount) {
		return domainwallet.DepositRequest{}, domainwallet.ErrDepositBelowMinimum
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
	provider := utils.NormalizeKey(req.Provider)
	if provider == domainwallet.DepositProviderQRIS || provider == domainwallet.DepositProviderQRISLY {
		qrisDeposit, err := s.prepareQRISDeposit(ctx, deposit)
		if err != nil {
			return domainwallet.DepositRequest{}, err
		}
		deposit = qrisDeposit
	}
	return s.Repo.CreateDepositRequest(ctx, deposit)
}

func (s *WalletService) GetMyDeposits(ctx context.Context, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	userID := strings.TrimSpace(authscope.FromContext(ctx).UserID)
	if userID == "" {
		return nil, 0, domainwallet.ErrInvalidAmount
	}
	return s.Repo.GetDeposits(ctx, userID, params)
}

func (s *WalletService) GetDeposits(ctx context.Context, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	return s.Repo.GetDeposits(ctx, "", params)
}

func (s *WalletService) GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	return s.Repo.GetDepositWithUserByID(ctx, strings.TrimSpace(id))
}

func (s *WalletService) GetMyDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	userID := strings.TrimSpace(authscope.FromContext(ctx).UserID)
	if userID == "" {
		return domainwallet.DepositRequest{}, domainwallet.ErrInvalidAmount
	}
	deposit, err := s.Repo.GetDepositByID(ctx, id)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	if deposit.UserID != userID {
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
	if !utils.GetEnv("PAYMENT_WEBHOOK_ENABLED", false) {
		return domainwallet.DepositRequest{}, domainwallet.ErrPaymentWebhookDisabled
	}
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

func (s *WalletService) prepareQRISDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	if s.QRISProvider == nil {
		return domainwallet.DepositRequest{}, domainwallet.ErrQRISProviderUnavailable
	}

	feePercent := "5"
	if s.ConfigService != nil {
		if value, err := s.ConfigService.GetString(ctx, "payment.qris.fee_percent", feePercent); err != nil {
			return domainwallet.DepositRequest{}, err
		} else {
			feePercent = value
		}
	}

	_, amountWithFee, feeAmount, err := money.MarkupAmount(deposit.Amount, feePercent)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	amountWithFeeRupiah, err := wholeRupiahAmount(amountWithFee)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	if amountWithFeeRupiah < 1000 {
		return domainwallet.DepositRequest{}, domainwallet.ErrInvalidAmount
	}

	qrisResponse, err := s.QRISProvider.GenerateQRIS(ctx, qrisly.GenerateQRISRequest{
		Amount:       amountWithFeeRupiah,
		UniqueAmount: true,
	})
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}

	payableRupiah := qrisResponse.Data.PayableAmount()
	if payableRupiah <= 0 {
		payableRupiah = amountWithFeeRupiah
	}
	payableAmount := money.FormatCents(payableRupiah * 100)
	uniqueCodeValue := payableRupiah - amountWithFeeRupiah
	if uniqueCodeValue < 0 {
		uniqueCodeValue = 0
	}
	uniqueCodeAmount := money.FormatCents(uniqueCodeValue * 100)
	uniqueCode := fmt.Sprintf("%d", uniqueCodeValue)

	paymentReference := strings.TrimSpace(qrisResponse.Data.HistoryID.String())
	if paymentReference == "" {
		paymentReference = "QRISLY-" + strings.ToUpper(strings.ReplaceAll(deposit.Id, "-", "")[:12])
	}
	expiredAt := parseGatewayTime(qrisResponse.Data.ExpiryValue())

	deposit.Method = domainwallet.DepositMethodPaymentGateway
	deposit.Provider = domainwallet.DepositProviderQRISLY
	deposit.PaymentReference = paymentReference
	deposit.PaymentURL = strings.TrimSpace(qrisResponse.Data.ImageValue())
	deposit.ExpiredAt = expiredAt
	deposit.Metadata = utils.MustJSON(map[string]string{
		"payment_channel":    domainwallet.DepositProviderQRISLY,
		"credit_amount":      deposit.Amount,
		"fee_percent":        feePercent,
		"fee_amount":         feeAmount,
		"unique_code":        uniqueCode,
		"unique_code_amount": uniqueCodeAmount,
		"payable_amount":     payableAmount,
		"qris_base_amount":   amountWithFee,
		"qris_image_url":     qrisResponse.Data.ImageValue(),
		"qris_string":        qrisResponse.Data.QRISString,
		"qris_merchant_name": qrisResponse.Data.MerchantValue(),
		"qrisly_history_id":  qrisResponse.Data.HistoryID.String(),
		"qrisly_qris_id":     qrisResponse.Data.QRISID.String(),
		"qrisly_status":      qrisResponse.Data.PaymentStatus,
		"qrisly_expiry_time": qrisResponse.Data.ExpiryValue(),
	})
	return deposit, nil
}

func wholeRupiahAmount(value string) (int64, error) {
	cents, err := money.ParseCents(value)
	if err != nil || cents <= 0 || cents%100 != 0 {
		return 0, domainwallet.ErrInvalidAmount
	}
	return cents / 100, nil
}

func belowMinimumDeposit(amount string) bool {
	amountCents, err := money.ParseCents(amount)
	if err != nil {
		return true
	}
	minimumCents, err := money.ParseCents(domainwallet.DepositMinimumAmount)
	if err != nil {
		return true
	}
	return amountCents < minimumCents
}

func parseGatewayTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05-07:00",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}
	return nil
}
