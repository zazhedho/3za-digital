package serviceprovider

import (
	"context"
	"errors"
	"time"

	domaincatalog "3za-digital/internal/domain/catalog"
	domainprovider "3za-digital/internal/domain/provider"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/money"
	"3za-digital/utils"
)

var (
	ErrProviderClientUnavailable = errors.New("provider client unavailable")
	ErrProviderEmptyResponse     = errors.New("provider returned empty response")
)

type ProviderService struct {
	Repo            interfaceprovider.RepoProviderInterface
	ProviderFactory interfaceprovider.ClientFactory
}

func NewProviderService(repo interfaceprovider.RepoProviderInterface, providerFactory interfaceprovider.ClientFactory) *ProviderService {
	return &ProviderService{
		Repo:            repo,
		ProviderFactory: providerFactory,
	}
}

func (s *ProviderService) GetH2HBalance(ctx context.Context) (domainprovider.BalanceSnapshot, error) {
	if s.ProviderFactory == nil {
		return domainprovider.BalanceSnapshot{}, ErrProviderClientUnavailable
	}

	client, err := s.ProviderFactory()
	if err != nil {
		return domainprovider.BalanceSnapshot{}, err
	}

	balance, err := client.GetBalance(ctx)
	if err != nil {
		return domainprovider.BalanceSnapshot{}, err
	}
	if balance == nil {
		return domainprovider.BalanceSnapshot{}, ErrProviderEmptyResponse
	}

	raw := balance.Raw
	if len(raw) == 0 {
		raw = utils.MustJSON(balance)
	}

	snapshot := domainprovider.BalanceSnapshot{
		Id:          utils.CreateUUID(),
		Provider:    domaincatalog.ProviderH2H,
		Balance:     money.NormalizeOrZero(balance.Balance.String()),
		RawResponse: raw,
		CreatedAt:   time.Now(),
	}
	if err := s.Repo.StoreBalanceSnapshot(ctx, snapshot); err != nil {
		return domainprovider.BalanceSnapshot{}, err
	}

	return snapshot, nil
}

func (s *ProviderService) GetAPILogs(ctx context.Context, params filter.BaseParams) ([]domainprovider.APILog, int64, error) {
	return s.Repo.GetAPILogs(ctx, params)
}

var _ interfaceprovider.ServiceProviderInterface = (*ProviderService)(nil)
