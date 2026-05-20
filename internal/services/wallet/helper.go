package servicewallet

import (
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	"3za-digital/utils"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"gorm.io/gorm"
)

func verifyWebhookSignature(provider string, req dto.PaymentWebhookRequest) error {
	secret := utils.GetEnv("PAYMENT_WEBHOOK_SECRET_"+utils.EnvKey(provider), "")
	if secret == "" {
		return nil
	}

	signature := strings.TrimSpace(req.Signature)
	if signature == "" {
		return domainwallet.ErrInvalidSignature
	}
	message := strings.Join([]string{
		strings.TrimSpace(provider),
		strings.TrimSpace(req.PaymentReference),
		strings.TrimSpace(req.Amount),
		strings.TrimSpace(req.Status),
	}, "|")

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(signature)), []byte(expected)) {
		return domainwallet.ErrInvalidSignature
	}
	return nil
}

func isDepositWebhookStatus(status string) bool {
	switch status {
	case domainwallet.DepositStatusPaid,
		domainwallet.DepositStatusExpired,
		domainwallet.DepositStatusFailed,
		domainwallet.DepositStatusCancelled:
		return true
	default:
		return false
	}
}

func normalizeGatewayStatus(status string) string {
	status = utils.NormalizeKey(status)
	switch status {
	case "paid", "success", "settlement", "capture":
		return domainwallet.DepositStatusPaid
	case "expired":
		return domainwallet.DepositStatusExpired
	case "failed", "deny":
		return domainwallet.DepositStatusFailed
	case "cancel", "cancelled", "canceled": //nolint:misspell
		return domainwallet.DepositStatusCancelled
	default:
		return status
	}
}

func IsPublicError(err error) bool {
	return errors.Is(err, domainwallet.ErrInactiveWallet) ||
		errors.Is(err, domainwallet.ErrInsufficientBalance) ||
		errors.Is(err, domainwallet.ErrInvalidAmount) ||
		errors.Is(err, domainwallet.ErrInvalidDirection) ||
		errors.Is(err, domainwallet.ErrDepositAlreadyFinal) ||
		errors.Is(err, domainwallet.ErrDepositAmountMismatch) ||
		errors.Is(err, domainwallet.ErrInvalidSignature)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
