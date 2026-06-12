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
	QRISProviders       map[string]interfacewallet.QRISPaymentProvider
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
	if provider == nil {
		return s
	}
	if s.QRISProviders == nil {
		s.QRISProviders = map[string]interfacewallet.QRISPaymentProvider{}
	}
	s.QRISProviders[utils.NormalizeKey(provider.Provider())] = provider
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
	case domainwallet.DepositProviderQRISLY, domainwallet.DepositProviderBOQRIS:
		qrisDeposit, err := s.prepareDynamicQRISDeposit(ctx, deposit)
		if err != nil {
			return domainwallet.DepositRequest{}, err
		}
		deposit = qrisDeposit
	default:
		manualDeposit, err := s.prepareManualReviewDeposit(ctx, deposit)
		if err != nil {
			return domainwallet.DepositRequest{}, err
		}
		deposit = manualDeposit
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
	deposits, err = s.syncDynamicQRISDeposits(ctx, deposits)
	return deposits, total, err
}

func (s *WalletService) GetDeposits(ctx context.Context, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	deposits, total, err := s.Repo.GetDeposits(ctx, "", params)
	if err != nil {
		return nil, 0, err
	}
	deposits, err = s.syncDynamicQRISDeposits(ctx, deposits)
	return deposits, total, err
}

func (s *WalletService) GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	deposit, err := s.Repo.GetDepositWithUserByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	return s.syncDynamicQRISDeposit(ctx, deposit)
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
	return s.syncDynamicQRISDeposit(ctx, deposit)
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

func (s *WalletService) depositAmountParts(ctx context.Context, deposit domainwallet.DepositRequest) (depositAmountParts, error) {
	feePercent, err := s.topupFeePercent(ctx)
	if err != nil {
		return depositAmountParts{}, err
	}
	_, amountWithFee, feeAmount, err := money.MarkupAmount(deposit.Amount, feePercent)
	if err != nil {
		return depositAmountParts{}, err
	}
	amountWithFeeRupiah, err := wholeRupiahAmount(amountWithFee)
	if err != nil {
		return depositAmountParts{}, err
	}
	if amountWithFeeRupiah < 1000 {
		return depositAmountParts{}, domainwallet.ErrInvalidAmount
	}
	return depositAmountParts{
		FeePercent:          feePercent,
		AmountWithFee:       amountWithFee,
		FeeAmount:           feeAmount,
		AmountWithFeeRupiah: amountWithFeeRupiah,
	}, nil
}

func (s *WalletService) prepareStaticQRISDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	parts, err := s.depositAmountParts(ctx, deposit)
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
	qrisProvider, providerName, err := s.dynamicQRISProvider(ctx, deposit.Provider)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	if qrisProvider == nil {
		detail := s.QRISInitError
		if detail == nil {
			detail = domainwallet.ErrQRISProviderUnavailable
		}
		return domainwallet.DepositRequest{}, fmt.Errorf("%w: %w", domainwallet.ErrQRISProviderUnavailable, detail)
	}

	parts, err := s.depositAmountParts(ctx, deposit)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	invoiceNo := dynamicQRISInvoiceNo(deposit.Id)
	qrisResponse, err := qrisProvider.GenerateQRIS(ctx, interfacewallet.QRISGenerateRequest{
		Amount:    parts.AmountWithFeeRupiah,
		InvoiceNo: invoiceNo,
	})
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}

	payableRupiah := qrisResponse.PayableAmount
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

	paymentReference := strings.TrimSpace(qrisResponse.PaymentReference)
	if paymentReference == "" {
		paymentReference = strings.ToUpper(providerName) + "-" + strings.ToUpper(strings.ReplaceAll(deposit.Id, "-", "")[:12])
	}
	expiredAt := parseGatewayTime(qrisResponse.ExpiresAt)

	deposit.Method = domainwallet.DepositMethodPaymentGateway
	deposit.Provider = providerName
	deposit.PaymentReference = paymentReference
	deposit.PaymentURL = strings.TrimSpace(qrisResponse.QRISImageURL)
	deposit.ExpiredAt = expiredAt
	metadata := map[string]string{
		"payment_channel":     providerName,
		"qris_type":           "dynamic",
		"credit_amount":       deposit.Amount,
		"fee_percent":         parts.FeePercent,
		"fee_amount":          parts.FeeAmount,
		"unique_code":         uniqueCode,
		"unique_code_amount":  uniqueCodeAmount,
		"payable_amount":      payableAmount,
		"qris_base_amount":    parts.AmountWithFee,
		"qris_image_url":      qrisResponse.QRISImageURL,
		"qris_string":         qrisResponse.QRISString,
		"qris_merchant_name":  qrisResponse.MerchantName,
		"qris_provider":       providerName,
		"qris_transaction_id": qrisResponse.TransactionID,
		"qris_status":         qrisResponse.Status,
		"qris_expiry_time":    qrisResponse.ExpiresAt,
		"invoice_no":          invoiceNo,
	}
	switch providerName {
	case domainwallet.DepositProviderQRISLY:
		metadata["qrisly_history_id"] = qrisResponse.PaymentReference
		metadata["qrisly_status"] = qrisResponse.Status
		metadata["qrisly_expiry_time"] = qrisResponse.ExpiresAt
	case domainwallet.DepositProviderBOQRIS:
		metadata["boqris_transaction_id"] = qrisResponse.TransactionID
		metadata["boqris_status"] = qrisResponse.Status
		metadata["boqris_expiry_time"] = qrisResponse.ExpiresAt
	}
	deposit.Metadata = utils.MustJSON(metadata)
	return deposit, nil
}

func (s *WalletService) prepareManualReviewDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	parts, err := s.depositAmountParts(ctx, deposit)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}
	if deposit.Provider == "" {
		deposit.Provider = "manual"
	}
	if deposit.PaymentReference == "" {
		deposit.PaymentReference = "MAN-" + strings.ToUpper(strings.ReplaceAll(deposit.Id, "-", "")[:12])
	}
	deposit.Method = domainwallet.DepositMethodManualAdmin
	deposit.Metadata = utils.MustJSON(map[string]string{
		"payment_channel":     "manual",
		"credit_amount":       deposit.Amount,
		"fee_percent":         parts.FeePercent,
		"fee_amount":          parts.FeeAmount,
		"payable_amount":      parts.AmountWithFee,
		"payment_instruction": "manual_review",
	})
	return deposit, nil
}

func (s *WalletService) topupFeePercent(ctx context.Context) (string, error) {
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

type depositAmountParts struct {
	FeePercent          string
	AmountWithFee       string
	FeeAmount           string
	AmountWithFeeRupiah int64
}

func (s *WalletService) syncDynamicQRISDeposits(ctx context.Context, deposits []domainwallet.DepositRequest) ([]domainwallet.DepositRequest, error) {
	for index := range deposits {
		updated, err := s.syncDynamicQRISDeposit(ctx, deposits[index])
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

func (s *WalletService) syncDynamicQRISDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	if !shouldSyncDynamicQRISDeposit(deposit) {
		return deposit, nil
	}
	if isDepositExpired(deposit) {
		return s.expireDynamicQRISDeposit(ctx, deposit)
	}

	providerName := utils.NormalizeKey(deposit.Provider)
	qrisProvider := s.QRISProviders[providerName]
	if qrisProvider == nil {
		return deposit, nil
	}

	paymentReference := dynamicQRISPaymentReference(deposit)
	if paymentReference == "" {
		return deposit, nil
	}
	statusResp, err := qrisProvider.GetPaymentStatus(ctx, paymentReference)
	if err != nil {
		return domainwallet.DepositRequest{}, err
	}

	gatewayStatus := utils.NormalizeKey(statusResp.Status)
	if gatewayStatus == "" {
		return deposit, nil
	}
	deposit.Metadata = mergeDepositMetadata(deposit.Metadata, dynamicQRISStatusMetadata(providerName, gatewayStatus))

	status := normalizeGatewayStatus(gatewayStatus)
	if status == domainwallet.DepositStatusPending {
		return deposit, nil
	}

	log := dynamicQRISStatusLog(deposit, providerName, gatewayStatus, paymentReference, statusResp)
	if status == domainwallet.DepositStatusPaid {
		updated, err := s.Repo.CompleteDepositByPaymentReference(ctx, providerName, deposit.PaymentReference, dynamicQRISPaidAmount(deposit, statusResp), log)
		return mergeSyncedDeposit(deposit, updated), err
	}
	if isDepositWebhookStatus(status) {
		updated, err := s.Repo.UpdateDepositStatusByPaymentReference(ctx, providerName, deposit.PaymentReference, status, log)
		return mergeSyncedDeposit(deposit, updated), err
	}
	return deposit, nil
}

func shouldSyncDynamicQRISDeposit(deposit domainwallet.DepositRequest) bool {
	provider := utils.NormalizeKey(deposit.Provider)
	return (provider == domainwallet.DepositProviderQRISLY || provider == domainwallet.DepositProviderBOQRIS) &&
		deposit.Status == domainwallet.DepositStatusPending &&
		strings.TrimSpace(deposit.PaymentReference) != ""
}

func dynamicQRISPaymentReference(deposit domainwallet.DepositRequest) string {
	metadata := depositMetadata(deposit.Metadata)
	if historyID := strings.TrimSpace(metadata["qrisly_history_id"]); historyID != "" {
		return historyID
	}
	if transactionID := strings.TrimSpace(metadata["boqris_transaction_id"]); transactionID != "" {
		return transactionID
	}
	if transactionID := strings.TrimSpace(metadata["qris_transaction_id"]); transactionID != "" {
		return transactionID
	}
	return strings.TrimSpace(deposit.PaymentReference)
}

func dynamicQRISPaidAmount(deposit domainwallet.DepositRequest, response *interfacewallet.QRISPaymentStatusResponse) string {
	if response != nil && response.Amount > 0 {
		return money.FormatCents(response.Amount * 100)
	}
	metadata := depositMetadata(deposit.Metadata)
	if payableAmount := strings.TrimSpace(metadata["payable_amount"]); payableAmount != "" {
		return payableAmount
	}
	return deposit.Amount
}

func dynamicQRISStatusLog(deposit domainwallet.DepositRequest, providerName string, status string, paymentReference string, statusResp *interfacewallet.QRISPaymentStatusResponse) domainwallet.PaymentGatewayLog {
	var payload json.RawMessage
	if statusResp != nil {
		payload = statusResp.Raw
	}
	if len(payload) == 0 {
		payload = utils.MustJSON(map[string]string{
			"payment_ref":    paymentReference,
			"payment_status": status,
			"amount":         dynamicQRISPaidAmount(deposit, statusResp),
		})
	}
	return domainwallet.PaymentGatewayLog{
		Id:               utils.CreateUUID(),
		Provider:         providerName,
		EventType:        "payment_status",
		RequestID:        paymentReference,
		PaymentReference: strings.TrimSpace(deposit.PaymentReference),
		Status:           status,
		Payload:          payload,
	}
}

func (s *WalletService) dynamicQRISProvider(ctx context.Context, requestedProvider string) (interfacewallet.QRISPaymentProvider, string, error) {
	providerName := utils.NormalizeKey(requestedProvider)
	if providerName == "" || providerName == domainwallet.DepositProviderQRISLY {
		providerName = defaultDynamicQRISProvider
		if s.ConfigService != nil {
			value, err := s.ConfigService.GetString(ctx, "payment.qris.dynamic_provider", defaultDynamicQRISProvider)
			if err != nil {
				return nil, "", err
			}
			if normalized := utils.NormalizeKey(value); normalized != "" {
				providerName = normalized
			}
		}
	}
	return s.QRISProviders[providerName], providerName, nil
}

func dynamicQRISInvoiceNo(depositID string) string {
	compactID := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(depositID), "-", ""))
	if len(compactID) > 12 {
		compactID = compactID[:12]
	}
	return "DEP" + compactID
}

func isDepositExpired(deposit domainwallet.DepositRequest) bool {
	return deposit.ExpiredAt != nil && time.Now().After(*deposit.ExpiredAt)
}

func (s *WalletService) expireDynamicQRISDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	providerName := utils.NormalizeKey(deposit.Provider)
	paymentReference := dynamicQRISPaymentReference(deposit)
	deposit.Metadata = mergeDepositMetadata(deposit.Metadata, dynamicQRISStatusMetadata(providerName, domainwallet.DepositStatusExpired))
	log := dynamicQRISStatusLog(deposit, providerName, domainwallet.DepositStatusExpired, paymentReference, nil)
	updated, err := s.Repo.UpdateDepositStatusByPaymentReference(ctx, providerName, deposit.PaymentReference, domainwallet.DepositStatusExpired, log)
	return mergeSyncedDeposit(deposit, updated), err
}

func dynamicQRISStatusMetadata(providerName string, status string) map[string]string {
	metadata := map[string]string{
		"qris_status":     status,
		"qris_checked_at": time.Now().Format(time.RFC3339),
	}
	switch providerName {
	case domainwallet.DepositProviderQRISLY:
		metadata["qrisly_status"] = status
		metadata["qrisly_checked_at"] = metadata["qris_checked_at"]
	case domainwallet.DepositProviderBOQRIS:
		metadata["boqris_status"] = status
		metadata["boqris_checked_at"] = metadata["qris_checked_at"]
	}
	return metadata
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

const (
	defaultQRISFeePercent      = "5"
	defaultDynamicQRISProvider = domainwallet.DepositProviderQRISLY
)

func parseGatewayTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}
	if location, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, location)
		if err == nil {
			return &parsed
		}
	}
	return nil
}
