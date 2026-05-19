package handlerdashboard

import (
	"fmt"
	"net/http"

	domaincatalog "3za-digital/internal/domain/catalog"
	interfacedashboard "3za-digital/internal/interfaces/dashboard"
	"3za-digital/pkg/logger"
	"3za-digital/pkg/response"
	"3za-digital/utils"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	Service interfacedashboard.ServiceDashboardInterface
}

func NewDashboardHandler(service interfacedashboard.ServiceDashboardInterface) *DashboardHandler {
	return &DashboardHandler{Service: service}
}

func (h *DashboardHandler) GetSummary(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[DashboardHandler][GetSummary]"
	reqCtx := ctx.Request.Context()

	productType := ctx.DefaultQuery("product_type", domaincatalog.ProductTypeSMM)
	data, err := h.Service.GetSummary(reqCtx, productType)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.GetSummary; Error: %+v", logPrefix, err))
		res := response.InternalServerError(logID)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.Response(http.StatusOK, "Get dashboard summary successfully", logID, data)
	ctx.JSON(http.StatusOK, res)
}
