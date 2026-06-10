package handlerappconfig

import (
	domainaudit "3za-digital/internal/domain/audit"
	"3za-digital/internal/dto"
	interfaceappconfig "3za-digital/internal/interfaces/appconfig"
	interfaceaudit "3za-digital/internal/interfaces/audit"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/logger"
	"3za-digital/pkg/messages"
	"3za-digital/pkg/response"
	"3za-digital/utils"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppConfigHandler struct {
	Service      interfaceappconfig.ServiceAppConfigInterface
	AuditService interfaceaudit.ServiceAuditInterface
}

func NewAppConfigHandler(s interfaceappconfig.ServiceAppConfigInterface, auditService interfaceaudit.ServiceAuditInterface) *AppConfigHandler {
	return &AppConfigHandler{
		Service:      s,
		AuditService: auditService,
	}
}

func (h *AppConfigHandler) GetAll(ctx *gin.Context) {
	logId := utils.GenerateLogId(ctx)
	logPrefix := "[AppConfigHandler][GetAll]"
	reqCtx := ctx.Request.Context()

	params, err := filter.GetBaseParams(ctx, "category", "asc", 50)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; GetBaseParams; Error: %+v", logPrefix, err))
		res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logId, nil)
		res.Error = response.Errors{Code: http.StatusBadRequest, Message: "invalid query parameters"}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	params.Filters = filter.WhitelistStringFilter(params.Filters, []string{"category"})

	data, total, err := h.Service.GetAll(reqCtx, params)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.GetAll; Error: %+v", logPrefix, err))
		res := response.InternalServerError(logId)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.PaginationResponse(http.StatusOK, int(total), params.Page, params.Limit, logId, data)
	ctx.JSON(http.StatusOK, res)
}

func (h *AppConfigHandler) GetByID(ctx *gin.Context) {
	logId := utils.GenerateLogId(ctx)
	logPrefix := "[AppConfigHandler][GetByID]"
	reqCtx := ctx.Request.Context()

	id, err := utils.ValidateUUID(ctx, logId)
	if err != nil {
		return
	}

	data, err := h.Service.GetByID(reqCtx, id)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.GetByID; Error: %+v", logPrefix, err))
		res := response.Response(http.StatusNotFound, "Configuration not found", logId, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "configuration not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := response.Response(http.StatusOK, "Get configuration successfully", logId, data)
	ctx.JSON(http.StatusOK, res)
}

func (h *AppConfigHandler) Update(ctx *gin.Context) {
	logId := utils.GenerateLogId(ctx)
	logPrefix := "[AppConfigHandler][Update]"
	reqCtx := ctx.Request.Context()

	id, err := utils.ValidateUUID(ctx, logId)
	if err != nil {
		return
	}

	var req dto.UpdateAppConfig
	if err := ctx.BindJSON(&req); err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; BindJSON ERROR: %s;", logPrefix, err.Error()))
		res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logId, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	before, _ := h.Service.GetByID(reqCtx, id)
	data, err := h.Service.Update(reqCtx, id, req)
	if err != nil {
		h.writeAudit(ctx, domainaudit.AuditEvent{
			Action:       domainaudit.ActionUpdate,
			Resource:     "config",
			ResourceID:   id,
			Status:       domainaudit.StatusFailed,
			Message:      "Failed to update configuration",
			ErrorMessage: err.Error(),
			BeforeData:   before,
			AfterData:    req,
		})

		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.Update; Error: %+v", logPrefix, err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := response.Response(http.StatusNotFound, messages.NotFound, logId, nil)
			res.Error = "config data not found"
			ctx.JSON(http.StatusNotFound, res)
			return
		}

		res := response.InternalServerError(logId)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	h.writeAudit(ctx, domainaudit.AuditEvent{
		Action:     domainaudit.ActionUpdate,
		Resource:   "config",
		ResourceID: data.Id,
		Status:     domainaudit.StatusSuccess,
		Message:    "Updated configuration",
		BeforeData: before,
		AfterData:  data,
	})

	res := response.Response(http.StatusOK, "Configuration updated successfully", logId, data)
	ctx.JSON(http.StatusOK, res)
}

func (h *AppConfigHandler) GetSupportContact(ctx *gin.Context) {
	logId := utils.GenerateLogId(ctx)
	reqCtx := ctx.Request.Context()

	data, err := h.Service.GetSupportContact(reqCtx)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("[AppConfigHandler][GetSupportContact]; Service.GetSupportContact; Error: %+v", err))
		res := response.InternalServerError(logId)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response(http.StatusOK, "Get support contact successfully", logId, data)
	ctx.JSON(http.StatusOK, res)
}
