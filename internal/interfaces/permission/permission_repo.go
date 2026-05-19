package interfacepermission

import (
	domainpermission "3za-digital/internal/domain/permission"
	interfacegeneric "3za-digital/internal/interfaces/generic"
	"context"
)

type RepoPermissionInterface interface {
	interfacegeneric.GenericRepository[domainpermission.Permission]

	GetByName(ctx context.Context, name string) (domainpermission.Permission, error)
	GetByResource(ctx context.Context, resource string) ([]domainpermission.Permission, error)
	GetUserPermissions(ctx context.Context, userId string) ([]domainpermission.Permission, error)
}
