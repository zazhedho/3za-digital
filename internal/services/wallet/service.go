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
	QRISProvider        interfacewallet.QRISPaymentProvider
	QRISInitError       error
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

func (s *WalletService) WithQRISProvider(provider interfacewallet.QRISPaymentProvider) *WalletService {
	s.QRISProvider = provider
	return s
}

func (s *WalletService) WithQRISInitError(err error) *WalletService {
	s.QRISInitError = err
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

func (s *WalletService) GetDepositSettings(ctx context.Context) (dto.DepositSettingsResponse, error) {
	feePercent := defaultQRISFeePercent
	staticImageURL := ""
	if s.ConfigService != nil {
		value, err := s.ConfigService.GetString(ctx, "payment.qris.fee_percent", feePercent)
		if err != nil {
			return dto.DepositSettingsResponse{}, err
		}
		feePercent = strings.TrimSpace(value)

		staticImageURL, err = s.ConfigService.GetString(ctx, "payment.qris.image_url", "")
		if err != nil {
			return dto.DepositSettingsResponse{}, err
		}
		staticImageURL = strings.TrimSpace(staticImageURL)
	}

	return dto.DepositSettingsResponse{
		MinimumAmount:      domainwallet.DepositMinimumAmount,
		QRISFeePercent:     feePercent,
		QRISStaticImageURL: staticImageURL,
	}, nil
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
	switch provider {
	case domainwallet.DepositProviderQRIS:
		qrisDeposit, err := s.prepareStaticQRISDeposit(ctx, deposit)
		if err != nil {
			return domainwallet.DepositRequest{}, err
		}
		deposit = qrisDeposit
	case domainwallet.DepositProviderQRISLY:
		qrisDeposit, err := s.prepareDynamicQRISDeposit(ctx, deposit)
		if err != nil {
			return domainwallet.DepositRequest{}, err
		}
		deposit = qrisDeposit
	default:
		// Fallback for manual review or unknown providers
		if deposit.Provider == "" {
			deposit.Provider = "manual"
		}
		if deposit.PaymentReference == "" {
			deposit.PaymentReference = "MAN-" + strings.ToUpper(strings.ReplaceAll(deposit.Id, "-", "")[:12])
		}
	}
	return s.Repo.CreateDepositRequest(ctx, deposit)
}

func (s *WalletService) GetMyDeposits(ctx context.Context, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	userID := strings.TrimSpace(authscope.FromContext(ctx).UserID)
	if userID == "" {
		return nil, 0, domainwallet.ErrInvalidAmount
	}
	deposits, total, err := s.Repo.GetDeposits(ctx, userID, params)
	if err != nil {
		return nil, 0, err
	}
	deposits, err = s.syncQRISLYDeposits(ctx, deposits)
	return deposits, total, err
}

func (s *WalletService) GetDeposits(ctx context.Context, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	deposits, total, err := s.Repo.GetDeposits(ctx, "", params)
	if err != nil {
		return nil, 0, err
	}
	deposits, err = s.syncQRISLYDeposits(ctx, deposits)
	return deposits, total, err
}

func (s *WalletService) GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	deposit, err := s.Repo.GetDepositWithUserByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	return s.syncQRISLYDeposit(ctx, deposit)
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
	return s.syncQRISLYDeposit(ctx, deposit)
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

func (s *WalletService) AdminCancelDeposit(ctx context.Context, depositID string, req dto.AdminDepositCancelRequest) (domainwallet.DepositRequest, error) {
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
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return domainwallet.DepositRequest{}, domainwallet.ErrDepositCancelReasonRequired
	}
	return s.Repo.CancelDeposit(ctx, deposit.Id, actorID, reason)
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

func (s *WalletService) qrisAmountParts(ctx context.Context, deposit domainwallet.DepositRequest) (qrisAmountParts, error) {
	feePercent, err := s.qrisFeePercent(ctx)
	if err != nil {
		return qrisAmountParts{}, err
	}
	_, amountWithFee, feeAmount, err := money.MarkupAmount(deposit.Amount, feePercent)
	if err != nil {
		return qrisAmountParts{}, err
	}
	amountWithFeeRupiah, err := wholeRupiahAmount(amountWithFee)
	if err != nil {
		return qrisAmountParts{}, err
	}
	if amountWithFeeRupiah < 1000 {
		return qrisAmountParts{}, domainwallet.ErrInvalidAmount
	}
	return qrisAmountParts{
		FeePercent:          feePercent,
		AmountWithFee:       amountWithFee,
		FeeAmount:           feeAmount,
		AmountWithFeeRupiah: amountWithFeeRupiah,
	}, nil
}

func (s *WalletService) prepareStaticQRISDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	parts, err := s.qrisAmountParts(ctx, deposit)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}

	imageURL := ""
	merchantName := ""
	if s.ConfigService != nil {
		if imageURL, err = s.ConfigService.GetString(ctx, "payment.qris.image_url", ""); err != nil {
			return domainwallet.DepositRequest{}, err
		}
		if merchantName, err = s.ConfigService.GetString(ctx, "payment.qris.merchant_name", ""); err != nil {
			return domainwallet.DepositRequest{}, err
		}
	}
	imageURL = strings.TrimSpace(imageURL)
	merchantName = strings.TrimSpace(merchantName)
	if imageURL == "" {
		return domainwallet.DepositRequest{}, domainwallet.ErrQRISStaticImageUnavailable
	}

	uniqueCodeValue := staticQRISUniqueCode(deposit.Id)
	payableRupiah := parts.AmountWithFeeRupiah + uniqueCodeValue
	payableAmount := money.FormatCents(payableRupiah * 100)
	uniqueCodeAmount := money.FormatCents(uniqueCodeValue * 100)
	uniqueCode := fmt.Sprintf("%d", uniqueCodeValue)

	deposit.Method = domainwallet.DepositMethodPaymentGateway
	deposit.Provider = domainwallet.DepositProviderQRIS
	deposit.PaymentReference = "QRIS-" + strings.ToUpper(strings.ReplaceAll(deposit.Id, "-", "")[:12])
	deposit.PaymentURL = imageURL
	deposit.Metadata = utils.MustJSON(map[string]string{
		"payment_channel":     domainwallet.DepositProviderQRIS,
		"qris_type":           "static",
		"credit_amount":       deposit.Amount,
		"fee_percent":         parts.FeePercent,
		"fee_amount":          parts.FeeAmount,
		"unique_code":         uniqueCode,
		"unique_code_amount":  uniqueCodeAmount,
		"payable_amount":      payableAmount,
		"qris_base_amount":    parts.AmountWithFee,
		"qris_image_url":      imageURL,
		"qris_merchant_name":  merchantName,
		"payment_instruction": "pay_exact_amount",
	})
	return deposit, nil
}

func (s *WalletService) prepareDynamicQRISDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	if s.QRISProvider == nil {
		detail := s.QRISInitError
		if detail == nil {
			detail = domainwallet.ErrQRISProviderUnavailable
		}
		return domainwallet.DepositRequest{}, fmt.Errorf("%w: %w", domainwallet.ErrQRISProviderUnavailable, detail)
	}

	parts, err := s.qrisAmountParts(ctx, deposit)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	qrisResponse, err := s.QRISProvider.GenerateQRIS(ctx, qrisly.GenerateQRISRequest{
		Amount:       parts.AmountWithFeeRupiah,
		UniqueAmount: true,
	})
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}

	payableRupiah := qrisResponse.Data.PayableAmount()
	if payableRupiah <= 0 {
		payableRupiah = parts.AmountWithFeeRupiah
	}
	payableAmount := money.FormatCents(payableRupiah * 100)
	uniqueCodeValue := payableRupiah - parts.AmountWithFeeRupiah
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
		"qris_type":          "dynamic",
		"credit_amount":      deposit.Amount,
		"fee_percent":        parts.FeePercent,
		"fee_amount":         parts.FeeAmount,
		"unique_code":        uniqueCode,
		"unique_code_amount": uniqueCodeAmount,
		"payable_amount":     payableAmount,
		"qris_base_amount":   parts.AmountWithFee,
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

func (s *WalletService) qrisFeePercent(ctx context.Context) (string, error) {
	feePercent := defaultQRISFeePercent
	if s.ConfigService == nil {
		return feePercent, nil
	}
	value, err := s.ConfigService.GetString(ctx, "payment.qris.fee_percent", feePercent)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

type qrisAmountParts struct {
	FeePercent          string
	AmountWithFee       string
	FeeAmount           string
	AmountWithFeeRupiah int64
}

func (s *WalletService) syncQRISLYDeposits(ctx context.Context, deposits []domainwallet.DepositRequest) ([]domainwallet.DepositRequest, error) {
	for index := range deposits {
		updated, err := s.syncQRISLYDeposit(ctx, deposits[index])
		if err != nil {
			return nil, err
		}
		if updated.User == nil {
			updated.User = deposits[index].User
		}
		deposits[index] = updated
	}
	return deposits, nil
}

func (s *WalletService) syncQRISLYDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	if s.QRISProvider == nil || !shouldSyncQRISLYDeposit(deposit) {
		return deposit, nil
	}

	historyID := qrislyHistoryID(deposit)
	if historyID == "" {
		return deposit, nil
	}
	statusResp, err := s.QRISProvider.GetPaymentStatus(ctx, historyID)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}

	qrislyStatus := qrislyPaymentStatus(statusResp.Data)
	if qrislyStatus == "" {
		return deposit, nil
	}
	deposit.Metadata = mergeDepositMetadata(deposit.Metadata, map[string]string{
		"qrisly_status":     qrislyStatus,
		"qrisly_checked_at": time.Now().Format(time.RFC3339),
	})

	status := normalizeGatewayStatus(qrislyStatus)
	if status == domainwallet.DepositStatusPending {
		return deposit, nil
	}

	log := qrislyStatusLog(deposit, qrislyStatus, statusResp)
	if status == domainwallet.DepositStatusPaid {
		updated, err := s.Repo.CompleteDepositByPaymentReference(ctx, domainwallet.DepositProviderQRISLY, deposit.PaymentReference, qrislyPaidAmount(deposit, statusResp.Data), log)
		return mergeSyncedDeposit(deposit, updated), err
	}
	if isDepositWebhookStatus(status) {
		updated, err := s.Repo.UpdateDepositStatusByPaymentReference(ctx, domainwallet.DepositProviderQRISLY, deposit.PaymentReference, status, log)
		return mergeSyncedDeposit(deposit, updated), err
	}
	return deposit, nil
}

func shouldSyncQRISLYDeposit(deposit domainwallet.DepositRequest) bool {
	return utils.NormalizeKey(deposit.Provider) == domainwallet.DepositProviderQRISLY &&
		deposit.Status == domainwallet.DepositStatusPending &&
		strings.TrimSpace(deposit.PaymentReference) != ""
}

func qrislyHistoryID(deposit domainwallet.DepositRequest) string {
	metadata := depositMetadata(deposit.Metadata)
	if historyID := strings.TrimSpace(metadata["qrisly_history_id"]); historyID != "" {
		return historyID
	}
	return strings.TrimSpace(deposit.PaymentReference)
}

func qrislyPaymentStatus(data qrisly.PaymentStatusData) string {
	if status := utils.NormalizeKey(data.PaymentStatus); status != "" {
		return status
	}
	return utils.NormalizeKey(data.Status)
}

func qrislyPaidAmount(deposit domainwallet.DepositRequest, data qrisly.PaymentStatusData) string {
	if data.Amount.Int64() > 0 {
		return money.FormatCents(data.Amount.Int64() * 100)
	}
	metadata := depositMetadata(deposit.Metadata)
	if payableAmount := strings.TrimSpace(metadata["payable_amount"]); payableAmount != "" {
		return payableAmount
	}
	return deposit.Amount
}

func qrislyStatusLog(deposit domainwallet.DepositRequest, status string, statusResp *qrisly.PaymentStatusResponse) domainwallet.PaymentGatewayLog {
	payload := statusResp.Raw
	if len(payload) == 0 {
		payload = utils.MustJSON(map[string]string{
			"history_id":     qrislyHistoryID(deposit),
			"payment_status": status,
			"amount":         qrislyPaidAmount(deposit, statusResp.Data),
		})
	}
	return domainwallet.PaymentGatewayLog{
		Id:               utils.CreateUUID(),
		Provider:         domainwallet.DepositProviderQRISLY,
		EventType:        "payment_status",
		RequestID:        qrislyHistoryID(deposit),
		PaymentReference: strings.TrimSpace(deposit.PaymentReference),
		Status:           status,
		Payload:          payload,
	}
}

func mergeSyncedDeposit(original domainwallet.DepositRequest, updated domainwallet.DepositRequest) domainwallet.DepositRequest {
	if updated.Id == "" {
		return original
	}
	updated.Metadata = original.Metadata
	if updated.User == nil {
		updated.User = original.User
	}
	return updated
}

func mergeDepositMetadata(raw json.RawMessage, values map[string]string) json.RawMessage {
	metadata := depositMetadata(raw)
	for key, value := range values {
		metadata[key] = value
	}
	return utils.MustJSON(metadata)
}

func depositMetadata(raw json.RawMessage) map[string]string {
	metadata := map[string]string{}
	if len(raw) == 0 {
		return metadata
	}
	_ = json.Unmarshal(raw, &metadata)
	return metadata
}

func wholeRupiahAmount(value string) (int64, error) {
	cents, err := money.ParseCents(value)
	if err != nil || cents <= 0 || cents%100 != 0 {
		return 0, domainwallet.ErrInvalidAmount
	}
	return cents / 100, nil
}

func staticQRISUniqueCode(seed string) int64 {
	var sum int64
	for index, char := range strings.ReplaceAll(seed, "-", "") {
		sum += int64(index+1) * int64(char)
	}
	return (sum % 900) + 100
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

const defaultQRISFeePercent = "5"

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
