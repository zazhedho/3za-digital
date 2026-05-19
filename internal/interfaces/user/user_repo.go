package interfaceuser

import (
	domainuser "3za-digital/internal/domain/user"
	interfacegeneric "3za-digital/internal/interfaces/generic"
	"context"
)

type RepoUserInterface interface {
	interfacegeneric.GenericRepository[domainuser.Users]

	GetByEmail(ctx context.Context, email string) (domainuser.Users, error)
	GetByPhone(ctx context.Context, phone string) (domainuser.Users, error)
}
