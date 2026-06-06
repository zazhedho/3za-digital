package handlersmm

import (
	"errors"
	"fmt"
	"net/http"

	domainaudit "3za-digital/internal/domain/audit"
	domaincatalog "3za-digital/internal/domain/catalog"
	"3za-digital/internal/dto"
	handlercommon "3za-digital/internal/handlers/http/common"
	interfaceaudit "3za-digital/internal/interfaces/audit"
	interfacecatalog "3za-digital/internal/interfaces/catalog"
	interfaceorder "3za-digital/internal/interfaces/order"
	servicecatalog "3za-digital/internal/services/catalog"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/logger"
	"3za-digital/pkg/messages"
	"3za-digital/pkg/response"
	"3za-digital/utils"

	"github.com/gin-gonic/gin"
)

type SMMHandler struct {
	CatalogService interfacecatalog.ServiceCatalogInterface
	OrderService   interfaceorder.ServiceOrderInterface
	AuditService   interfaceaudit.ServiceAuditInterface
}

func NewSMMHandler(catalogService interfacecatalog.ServiceCatalogInterface, orderService interfaceorder.ServiceOrderInterface, auditServices ...interfaceaudit.ServiceAuditInterface) *SMMHandler {
	handler := &SMMHandler{
		CatalogService: catalogService,
		OrderService:   orderService,
	}
	if len(auditServices) > 0 {
		handler.AuditService = auditServices[0]
	}
	return handler
}

func (h *SMMHandler) writeAudit(ctx *gin.Context, event domainaudit.AuditEvent) {
	handlercommon.WriteAudit(ctx, h.AuditService, event, "SMMHandler")
}

func syncCatalogAuditPayload(req dto.SyncCatalogRequest, data dto.SyncCatalogResponse) map[string]interface{} {
	return map[string]interface{}{
		"provider":     data.Provider,
		"product_type": data.ProductType,
		"total":        data.Total,
		"synced":       data.Synced,
		"platform":     req.Platform,
		"category":     req.Category,
	}
}

func (h *SMMHandler) GetServices(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[SMMHandler][GetServices]"
	reqCtx := ctx.Request.Context()

	params, err := filter.GetBaseParams(ctx, "name", "asc", 50)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; GetBaseParams; Error: %+v", logPrefix, err))
		res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logID, nil)
		res.Error = response.Errors{Code: http.StatusBadRequest, Message: "invalid query parameters"}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	params.Filters = filter.WhitelistStringFilter(params.Filters, []string{
		"provider_service_id",
		"category",
		"brand",
		"platform",
		"is_active",
	})
	params.Filters["provider"] = domaincatalog.ProviderH2H
	params.Filters["product_type"] = domaincatalog.ProductTypeSMM

	data, total, err := h.CatalogService.GetAll(reqCtx, params)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.GetAll; Error: %+v", logPrefix, err))
		res := response.InternalServerError(logID)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.PaginationResponse(http.StatusOK, int(total), params.Page, params.Limit, logID, data)
	ctx.JSON(http.StatusOK, res)
}

func (h *SMMHandler) SyncServices(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[SMMHandler][SyncServices]"
	reqCtx := ctx.Request.Context()

	var req dto.SyncCatalogRequest
	if err := ctx.ShouldBind(&req); err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; ShouldBind; Error: %+v", logPrefix, err))
		res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logID, nil)
		res.Error = response.Errors{Code: http.StatusBadRequest, Message: "invalid request body"}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	data, err := h.CatalogService.Sync(reqCtx, domaincatalog.ProductTypeSMM, req)
	if err != nil {
		h.writeAudit(ctx, domainaudit.AuditEvent{
			Action:       domainaudit.ActionUpdate,
			Resource:     "smm_service_catalog",
			Status:       domainaudit.StatusFailed,
			Message:      "Failed to sync SMM service catalog",
			ErrorMessage: err.Error(),
			AfterData: map[string]interface{}{
				"product_type": domaincatalog.ProductTypeSMM,
				"platform":     req.Platform,
				"category":     req.Category,
			},
		})
		statusCode := http.StatusBadGateway
		publicMessage := "failed to sync SMM services from provider"
		if errors.Is(err, servicecatalog.ErrUnsupportedProductType) {
			statusCode = http.StatusBadRequest
			publicMessage = err.Error()
		}
		if errors.Is(err, servicecatalog.ErrProviderUnavailable) {
			statusCode = http.StatusServiceUnavailable
			publicMessage = err.Error()
		}

		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.Sync; Error: %+v", logPrefix, err))
		res := response.Response(statusCode, messages.MsgSomethingWrong, logID, nil)
		res.Error = response.Errors{Code: statusCode, Message: publicMessage}
		ctx.JSON(statusCode, res)
		return
	}

	h.writeAudit(ctx, domainaudit.AuditEvent{
		Action:    domainaudit.ActionUpdate,
		Resource:  "smm_service_catalog",
		Status:    domainaudit.StatusSuccess,
		Message:   "Synced SMM service catalog",
		AfterData: syncCatalogAuditPayload(req, data),
	})
	res := response.Response(http.StatusOK, "SMM services synced successfully", logID, data)
	ctx.JSON(http.StatusOK, res)
}
