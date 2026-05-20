package handlercommon

import (
	"fmt"

	"3za-digital/internal/authscope"
	domainaudit "3za-digital/internal/domain/audit"
	interfaceaudit "3za-digital/internal/interfaces/audit"
	"3za-digital/pkg/logger"
	"3za-digital/utils"

	"github.com/gin-gonic/gin"
)

func WriteAudit(ctx *gin.Context, auditService interfaceaudit.ServiceAuditInterface, event domainaudit.AuditEvent, scope string) {
	if auditService == nil {
		return
	}

	scopeData := authscope.FromContext(ctx.Request.Context())
	if event.ActorUserID == "" && scopeData.ActorUserID() != "" {
		event.ActorUserID = scopeData.ActorUserID()
	}
	if event.ActorRole == "" && scopeData.ActorRole() != "" {
		event.ActorRole = scopeData.ActorRole()
	}
	event.RequestID = utils.GetRequestID(ctx)
	event.IPAddress = ctx.ClientIP()
	event.UserAgent = ctx.GetHeader("User-Agent")
	event.Metadata = utils.MergeMetadata(event.Metadata, utils.GetImpersonationMetadata(ctx))

	if err := auditService.Store(ctx.Request.Context(), event); err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelWarn, fmt.Sprintf("[%s][Audit]; failed to store audit trail: %v", scope, err))
	}
}
