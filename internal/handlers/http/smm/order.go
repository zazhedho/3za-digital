package handlersmm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"3za-digital/internal/authscope"
	domaincatalog "3za-digital/internal/domain/catalog"
	"3za-digital/internal/dto"
	serviceorder "3za-digital/internal/services/order"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/logger"
	"3za-digital/pkg/messages"
	"3za-digital/pkg/response"
	"3za-digital/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (h *SMMHandler) GetOrders(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[SMMHandler][GetOrders]"
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
		"ref_id",
		"provider_service_id",
		"status",
		"created_by",
	})
	params.Filters["provider"] = domaincatalog.ProviderH2H
	params.Filters["product_type"] = domaincatalog.ProductTypeSMM
	scope := authscope.FromContext(reqCtx)
	if !canViewAllSMMOrders(scope) {
		params.Filters["created_by"] = scope.UserID
	}

	data, total, err := h.OrderService.GetAll(reqCtx, params)
	if err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Service.GetAll; Error: %+v", logPrefix, err))
		res := response.InternalServerError(logID)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := response.PaginationResponse(http.StatusOK, int(total), params.Page, params.Limit, logID, data)
	ctx.JSON(http.StatusOK, res)
}

func (h *SMMHandler) GetOrderByID(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[SMMHandler][GetOrderByID]"
	reqCtx := ctx.Request.Context()

	id, err := utils.ValidateUUID(ctx, logID)
	if err != nil {
		return
	}

	data, err := h.OrderService.GetByID(reqCtx, id)
	if err != nil {
		writeOrderError(ctx, logPrefix, logID, err)
		return
	}
	if data.ProductType != domaincatalog.ProductTypeSMM {
		res := response.Response(http.StatusNotFound, messages.MsgNotFound, logID, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "SMM order not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}
	if !canAccessSMMOrder(reqCtx, data.CreatedBy) {
		res := response.Response(http.StatusNotFound, messages.MsgNotFound, logID, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "SMM order not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := response.Response(http.StatusOK, "Get SMM order successfully", logID, data)
	ctx.JSON(http.StatusOK, res)
}

func (h *SMMHandler) CreateOrder(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[SMMHandler][CreateOrder]"
	reqCtx := ctx.Request.Context()

	var req dto.CreateOrderRequest
	if err := ctx.BindJSON(&req); err != nil {
		logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; BindJSON; Error: %+v", logPrefix, err))
		res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logID, nil)
		res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	data, err := h.OrderService.CreateOrder(reqCtx, domaincatalog.ProductTypeSMM, req)
	if err != nil {
		writeOrderError(ctx, logPrefix, logID, err)
		return
	}

	res := response.Response(http.StatusCreated, "SMM order created successfully", logID, data)
	ctx.JSON(http.StatusCreated, res)
}

func (h *SMMHandler) RefreshOrderStatus(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[SMMHandler][RefreshOrderStatus]"
	reqCtx := ctx.Request.Context()

	id, err := utils.ValidateUUID(ctx, logID)
	if err != nil {
		return
	}

	existing, err := h.OrderService.GetByID(reqCtx, id)
	if err != nil {
		writeOrderError(ctx, logPrefix, logID, err)
		return
	}
	if existing.ProductType != domaincatalog.ProductTypeSMM {
		res := response.Response(http.StatusNotFound, messages.MsgNotFound, logID, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "SMM order not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}
	if !canAccessSMMOrder(reqCtx, existing.CreatedBy) {
		res := response.Response(http.StatusNotFound, messages.MsgNotFound, logID, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "SMM order not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	data, err := h.OrderService.RefreshStatus(reqCtx, id)
	if err != nil {
		writeOrderError(ctx, logPrefix, logID, err)
		return
	}

	res := response.Response(http.StatusOK, "SMM order status refreshed successfully", logID, data)
	ctx.JSON(http.StatusOK, res)
}

func (h *SMMHandler) GetOrderStatusLogs(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	logPrefix := "[SMMHandler][GetOrderStatusLogs]"
	reqCtx := ctx.Request.Context()

	id, err := utils.ValidateUUID(ctx, logID)
	if err != nil {
		return
	}

	order, err := h.OrderService.GetByID(reqCtx, id)
	if err != nil {
		writeOrderError(ctx, logPrefix, logID, err)
		return
	}
	if order.ProductType != domaincatalog.ProductTypeSMM {
		res := response.Response(http.StatusNotFound, messages.MsgNotFound, logID, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "SMM order not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}
	if !canAccessSMMOrder(reqCtx, order.CreatedBy) {
		res := response.Response(http.StatusNotFound, messages.MsgNotFound, logID, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "SMM order not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	data, err := h.OrderService.GetStatusLogs(reqCtx, id)
	if err != nil {
		writeOrderError(ctx, logPrefix, logID, err)
		return
	}

	res := response.Response(http.StatusOK, "Get SMM order status logs successfully", logID, data)
	ctx.JSON(http.StatusOK, res)
}

func canAccessSMMOrder(ctx context.Context, createdBy *string) bool {
	scope := authscope.FromContext(ctx)
	if canViewAllSMMOrders(scope) {
		return true
	}
	if createdBy == nil {
		return false
	}
	return *createdBy == scope.UserID
}

func canViewAllSMMOrders(scope authscope.Scope) bool {
	return scope.Has("smm_orders", "list_all")
}

func writeOrderError(ctx *gin.Context, logPrefix string, logID uuid.UUID, err error) {
	logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Error: %+v", logPrefix, err))

	if errors.Is(err, gorm.ErrRecordNotFound) || serviceorder.IsNotFound(err) {
		res := response.Response(http.StatusNotFound, messages.MsgNotFound, logID, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "SMM order not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	if errors.Is(err, serviceorder.ErrOrderAlreadyFinal) {
		res := response.Response(http.StatusConflict, messages.MsgSomethingWrong, logID, nil)
		res.Error = response.Errors{Code: http.StatusConflict, Message: err.Error()}
		ctx.JSON(http.StatusConflict, res)
		return
	}

	if serviceorder.IsPublicError(err) {
		res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logID, nil)
		res.Error = response.Errors{Code: http.StatusBadRequest, Message: err.Error()}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := response.Response(http.StatusBadGateway, messages.MsgSomethingWrong, logID, nil)
	res.Error = response.Errors{Code: http.StatusBadGateway, Message: "provider transaction failed"}
	ctx.JSON(http.StatusBadGateway, res)
}
