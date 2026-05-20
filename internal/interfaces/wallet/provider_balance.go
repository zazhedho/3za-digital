package interfacewallet

import (
	"context"

	domainprovider "3za-digital/internal/domain/provider"
)

type MainBalanceProvider interface {
	GetH2HBalance(ctx context.Context) (domainprovider.BalanceSnapshot, error)
}
