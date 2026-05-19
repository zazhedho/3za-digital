package interfaceaudit

import (
	domainaudit "3za-digital/internal/domain/audit"
	"3za-digital/internal/dto"
	"3za-digital/pkg/filter"
	"context"
)

type ServiceAuditInterface interface {
	Store(ctx context.Context, req domainaudit.AuditEvent) error
	GetAll(ctx context.Context, params filter.BaseParams) ([]dto.AuditTrailResponse, int64, error)
	GetByID(ctx context.Context, id string) (dto.AuditTrailResponse, error)
}
