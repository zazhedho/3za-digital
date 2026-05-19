package handlermenu

import (
	domainaudit "3za-digital/internal/domain/audit"
	handlercommon "3za-digital/internal/handlers/http/common"

	"github.com/gin-gonic/gin"
)

func (h *MenuHandler) writeAudit(ctx *gin.Context, event domainaudit.AuditEvent) {
	handlercommon.WriteAudit(ctx, h.AuditService, event, "MenuHandler")
}
