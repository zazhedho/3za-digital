package handlerprovider

import (
	"fmt"
	"net/http"

	interfaceprovider "3za-digital/internal/interfaces/provider"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/logger"
	"3za-digital/pkg/messages"
	"3za-digital/pkg/response"
	"3za-digital/utils"

	"github.com/gin-gonic/gin"
)

type ProviderHandler struct {
	Service interfaceprovider.ServiceProviderInterface
}

func NewProviderHandler(service interfaceprovider.ServiceProviderInterface) *ProviderHandler {
	return &ProviderHandler{Service: service}
}

func (h *ProviderHandler) GetH2HBalance(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[ProviderHandler][GetH2HBalance]"
	reqCtx := ctx.Request.Context()

	data, err := h.Service.GetH2HBalance(reqCtx)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.GetH2HBalance; Error: %+v", logPrefix, err))
		res := response.Response(http.StatusBadGateway, messages.MsgSomethingWrong, logID, nil)
		res.Error = response.Errors{Code: http.StatusBadGateway, Message: "failed to get provider balance"}
		ctx.JSON(http.StatusBadGateway, res)
		return
	}

	res := response.Response(http.StatusOK, "Get H2H balance successfully", logID, data)
	ctx.JSON(http.StatusOK, res)
}

func (h *ProviderHandler) GetAPILogs(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[ProviderHandler][GetAPILogs]"
	reqCtx := ctx.Request.Context()

	params, err := filter.GetBaseParams(ctx, "created_at", "desc", 50)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; GetBaseParams; Error: %+v", logPrefix, err))
		res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logID, nil)
		res.Error = response.Errors{Code: http.StatusBadRequest, Message: "invalid query parameters"}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	params.Filters = filter.WhitelistStringFilter(params.Filters, []string{
		"provider",
		"product_type",
		"endpoint",
		"request_ref",
		"response_status",
	})

	data, total, err := h.Service.GetAPILogs(reqCtx, params)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.GetAPILogs; Error: %+v", logPrefix, err))
		res := response.InternalServerError(logID)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.PaginationResponse(http.StatusOK, int(total), params.Page, params.Limit, logID, data)
	ctx.JSON(http.StatusOK, res)
}
