package domainwallet

import "errors"

var (
	ErrInactiveWallet        = errors.New("wallet is inactive")
	ErrInsufficientBalance   = errors.New("insufficient wallet balance")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInvalidDirection      = errors.New("invalid wallet transaction direction")
	ErrDepositAlreadyFinal   = errors.New("deposit already has final status")
	ErrDepositAmountMismatch = errors.New("deposit amount mismatch")
)
