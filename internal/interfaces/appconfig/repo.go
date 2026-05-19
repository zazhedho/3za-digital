package interfaceappconfig

import (
	domainappconfig "3za-digital/internal/domain/appconfig"
	interfacegeneric "3za-digital/internal/interfaces/generic"
	"context"
)

type RepoAppConfigInterface interface {
	interfacegeneric.GenericRepository[domainappconfig.AppConfig]

	GetByKey(ctx context.Context, configKey string) (domainappconfig.AppConfig, error)
}
