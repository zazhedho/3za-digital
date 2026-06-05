package domainwallet

import (
	"3za-digital/pkg/money"
	"errors"
)

var (
	ErrInactiveWallet          = errors.New("wallet is inactive")
	ErrInsufficientBalance     = errors.New("insufficient wallet balance")
	ErrInvalidAmount           = money.ErrInvalidAmount
	ErrInvalidDirection        = errors.New("invalid wallet transaction direction")
	ErrDepositAlreadyFinal     = errors.New("deposit already has final status")
	ErrDepositAmountMismatch   = errors.New("deposit amount mismatch")
	ErrDepositBelowMinimum     = errors.New("minimum deposit amount is 10000")
	ErrInvalidSignature        = errors.New("invalid payment webhook signature")
	ErrPaymentWebhookDisabled  = errors.New("payment webhook is disabled")
	ErrInsufficientMainBalance = errors.New("insufficient main provider balance")
	ErrMainBalanceUnavailable  = errors.New("main provider balance unavailable")
	ErrQRISProviderUnavailable = errors.New("qris payment provider unavailable")
)
