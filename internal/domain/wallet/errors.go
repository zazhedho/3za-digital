package domainwallet

import (
	"3za-digital/pkg/money"
	"errors"
)

var (
	ErrInactiveWallet        = errors.New("wallet is inactive")
	ErrInsufficientBalance   = errors.New("insufficient wallet balance")
	ErrInvalidAmount         = money.ErrInvalidAmount
	ErrInvalidDirection      = errors.New("invalid wallet transaction direction")
	ErrDepositAlreadyFinal   = errors.New("deposit already has final status")
	ErrDepositAmountMismatch = errors.New("deposit amount mismatch")
	ErrInvalidSignature      = errors.New("invalid payment webhook signature")
)
