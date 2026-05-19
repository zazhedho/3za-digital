package interfaceaudit

import (
	domainaudit "3za-digital/internal/domain/audit"
	interfacegeneric "3za-digital/internal/interfaces/generic"
)

type RepoAuditInterface interface {
	interfacegeneric.GenericRepository[domainaudit.AuditTrail]
}
