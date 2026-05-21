package handlerwallet

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"3za-digital/internal/authscope"
	domainaudit "3za-digital/internal/domain/audit"
	domainwallet "3za-digital/internal/domain/wallet"
	"3za-digital/internal/dto"
	handlercommon "3za-digital/internal/handlers/http/common"
	interfaceaudit "3za-digital/internal/interfaces/audit"
	interfacewallet "3za-digital/internal/interfaces/wallet"
	servicewallet "3za-digital/internal/services/wallet"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/logger"
	"3za-digital/pkg/messages"
	"3za-digital/pkg/response"
	"3za-digital/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletHandler struct {
	Service      interfacewallet.ServiceWalletInterface
	AuditService interfaceaudit.ServiceAuditInterface
}

func NewWalletHandler(service interfacewallet.ServiceWalletInterface, auditServices ...interfaceaudit.ServiceAuditInterface) *WalletHandler {
	handler := &WalletHandler{Service: service}
	if len(auditServices) > 0 {
		handler.AuditService = auditServices[0]
	}
	return handler
}

func (h *WalletHandler) GetMyWallet(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	scope := authscope.FromContext(ctx.Request.Context())
	if scope.UserID == "" {
		writeWalletError(ctx, "[WalletHandler][GetMyWallet]", logID, gorm.ErrRecordNotFound)
		return
	}

	data, err := h.Service.GetMyWallet(ctx.Request.Context(), scope.UserID)
	if err != nil {
		writeWalletError(ctx, "[WalletHandler][GetMyWallet]", logID, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Response(http.StatusOK, "Get wallet successfully", logID, data))
}

func (h *WalletHandler) GetMyTransactions(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	scope := authscope.FromContext(ctx.Request.Context())
	params, err := filter.GetBaseParams(ctx, "created_at", "desc", 50)
	if err != nil {
		writeBadRequest(ctx, logID, "invalid query parameters")
		return
	}
	params.Filters = filter.WhitelistStringFilter(params.Filters, []string{"type", "direction", "reference"})

	data, total, err := h.Service.GetMyTransactions(ctx.Request.Context(), scope.UserID, params)
	if err != nil {
		writeWalletError(ctx, "[WalletHandler][GetMyTransactions]", logID, err)
		return
	}
	ctx.JSON(http.StatusOK, response.PaginationResponse(http.StatusOK, int(total), params.Page, params.Limit, logID, data))
}

func (h *WalletHandler) GetWallets(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	params, err := filter.GetBaseParams(ctx, "created_at", "desc", 50)
	if err != nil {
		writeBadRequest(ctx, logID, "invalid query parameters")
		return
	}
	params.Filters = filter.WhitelistStringFilter(params.Filters, []string{"user_id", "currency", "is_active"})

	data, total, err := h.Service.GetWallets(ctx.Request.Context(), params)
	if err != nil {
		writeWalletError(ctx, "[WalletHandler][GetWallets]", logID, err)
		return
	}
	ctx.JSON(http.StatusOK, response.PaginationResponse(http.StatusOK, int(total), params.Page, params.Limit, logID, data))
}

func (h *WalletHandler) GetDeposits(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	params, err := filter.GetBaseParams(ctx, "created_at", "desc", 50)
	if err != nil {
		writeBadRequest(ctx, logID, "invalid query parameters")
		return
	}
	params.Filters = filter.WhitelistStringFilter(params.Filters, []string{"user_id", "status", "method", "provider", "payment_reference"})

	data, total, err := h.Service.GetDeposits(ctx.Request.Context(), params)
	if err != nil {
		writeWalletError(ctx, "[WalletHandler][GetDeposits]", logID, err)
		return
	}
	ctx.JSON(http.StatusOK, response.PaginationResponse(http.StatusOK, int(total), params.Page, params.Limit, logID, data))
}

func (h *WalletHandler) AdminTopup(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	userID := strings.TrimSpace(ctx.Param("user_id"))
	if _, err := uuid.Parse(userID); err != nil {
		writeBadRequest(ctx, logID, "invalid user id")
		return
	}

	var req dto.AdminWalletTopupRequest
	if err := ctx.BindJSON(&req); err != nil {
		writeBindError(ctx, logID, req, err)
		return
	}
	data, err := h.Service.AdminTopup(ctx.Request.Context(), userID, req)
	if err != nil {
		h.writeWalletAudit(ctx, domainaudit.ActionCreate, "wallet_topup", userID, domainaudit.StatusFailed, "Failed to topup wallet", err.Error(), nil)
		writeWalletError(ctx, "[WalletHandler][AdminTopup]", logID, err)
		return
	}
	h.writeWalletAudit(ctx, domainaudit.ActionCreate, "wallet_topup", data.Id, domainaudit.StatusSuccess, "Topup wallet", "", data)
	ctx.JSON(http.StatusCreated, response.Response(http.StatusCreated, "Wallet topup successfully", logID, data))
}

func (h *WalletHandler) AdminApproveDeposit(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	id, err := utils.ValidateUUID(ctx, logID)
	if err != nil {
		return
	}

	var req dto.AdminDepositApproveRequest
	if err := ctx.BindJSON(&req); err != nil {
		writeBindError(ctx, logID, req, err)
		return
	}
	data, err := h.Service.AdminApproveDeposit(ctx.Request.Context(), id, req)
	if err != nil {
		h.writeWalletAudit(ctx, domainaudit.ActionUpdate, "deposit_request", id, domainaudit.StatusFailed, "Failed to approve deposit request", err.Error(), nil)
		writeWalletError(ctx, "[WalletHandler][AdminApproveDeposit]", logID, err)
		return
	}
	h.writeWalletAudit(ctx, domainaudit.ActionUpdate, "deposit_request", data.Id, domainaudit.StatusSuccess, "Approved deposit request", "", data)
	ctx.JSON(http.StatusOK, response.Response(http.StatusOK, "Deposit request approved successfully", logID, data))
}

func (h *WalletHandler) AdminAdjust(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	userID := strings.TrimSpace(ctx.Param("user_id"))
	if _, err := uuid.Parse(userID); err != nil {
		writeBadRequest(ctx, logID, "invalid user id")
		return
	}

	var req dto.AdminWalletAdjustRequest
	if err := ctx.BindJSON(&req); err != nil {
		writeBindError(ctx, logID, req, err)
		return
	}
	data, err := h.Service.AdminAdjust(ctx.Request.Context(), userID, req)
	if err != nil {
		h.writeWalletAudit(ctx, domainaudit.ActionUpdate, "wallet_adjustment", userID, domainaudit.StatusFailed, "Failed to adjust wallet", err.Error(), nil)
		writeWalletError(ctx, "[WalletHandler][AdminAdjust]", logID, err)
		return
	}
	h.writeWalletAudit(ctx, domainaudit.ActionUpdate, "wallet_adjustment", data.Id, domainaudit.StatusSuccess, "Adjusted wallet", "", data)
	ctx.JSON(http.StatusCreated, response.Response(http.StatusCreated, "Wallet adjusted successfully", logID, data))
}

func (h *WalletHandler) CreateDeposit(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	scope := authscope.FromContext(ctx.Request.Context())
	var req dto.CreateDepositRequest
	if err := ctx.BindJSON(&req); err != nil {
		writeBindError(ctx, logID, req, err)
		return
	}

	data, err := h.Service.CreateDeposit(ctx.Request.Context(), scope.UserID, req)
	if err != nil {
		h.writeWalletAudit(ctx, domainaudit.ActionCreate, "deposit_request", scope.UserID, domainaudit.StatusFailed, "Failed to create deposit request", err.Error(), nil)
		writeWalletError(ctx, "[WalletHandler][CreateDeposit]", logID, err)
		return
	}
	h.writeWalletAudit(ctx, domainaudit.ActionCreate, "deposit_request", data.Id, domainaudit.StatusSuccess, "Created deposit request", "", data)
	ctx.JSON(http.StatusCreated, response.Response(http.StatusCreated, "Deposit request created successfully", logID, data))
}

func (h *WalletHandler) GetMyDeposits(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	params, err := filter.GetBaseParams(ctx, "created_at", "desc", 50)
	if err != nil {
		writeBadRequest(ctx, logID, "invalid query parameters")
		return
	}
	params.Filters = filter.WhitelistStringFilter(params.Filters, []string{"status", "method", "provider", "payment_reference"})

	data, total, err := h.Service.GetMyDeposits(ctx.Request.Context(), params)
	if err != nil {
		writeWalletError(ctx, "[WalletHandler][GetMyDeposits]", logID, err)
		return
	}
	ctx.JSON(http.StatusOK, response.PaginationResponse(http.StatusOK, int(total), params.Page, params.Limit, logID, data))
}

func (h *WalletHandler) GetMyDepositByID(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	id, err := utils.ValidateUUID(ctx, logID)
	if err != nil {
		return
	}

	data, err := h.Service.GetMyDepositByID(ctx.Request.Context(), id)
	if err != nil {
		writeWalletError(ctx, "[WalletHandler][GetMyDepositByID]", logID, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Response(http.StatusOK, "Get deposit successfully", logID, data))
}

func (h *WalletHandler) PaymentWebhook(ctx *gin.Context) {
	logID := utils.GenerateLogId(ctx)
	provider := strings.TrimSpace(ctx.Param("provider"))
	var req dto.PaymentWebhookRequest
	if err := ctx.BindJSON(&req); err != nil {
		writeBindError(ctx, logID, req, err)
		return
	}

	data, err := h.Service.HandlePaymentWebhook(ctx.Request.Context(), provider, req)
	if err != nil {
		h.writeWalletAudit(ctx, domainaudit.ActionUpdate, "payment_webhook", req.PaymentReference, domainaudit.StatusFailed, "Failed to process payment webhook", err.Error(), map[string]interface{}{
			"provider":          provider,
			"payment_reference": req.PaymentReference,
			"status":            req.Status,
		})
		writeWalletError(ctx, "[WalletHandler][PaymentWebhook]", logID, err)
		return
	}
	h.writeWalletAudit(ctx, domainaudit.ActionUpdate, "payment_webhook", data.Id, domainaudit.StatusSuccess, "Processed payment webhook", "", map[string]interface{}{
		"provider":          provider,
		"payment_reference": req.PaymentReference,
		"status":            req.Status,
		"deposit":           data,
	})
	ctx.JSON(http.StatusOK, response.Response(http.StatusOK, "Payment webhook processed successfully", logID, data))
}

func (h *WalletHandler) writeWalletAudit(ctx *gin.Context, action, resource, resourceID, status, message, errMessage string, after interface{}) {
	handlercommon.WriteAudit(ctx, h.AuditService, domainaudit.AuditEvent{
		Action:       action,
		Resource:     resource,
		ResourceID:   resourceID,
		Status:       status,
		Message:      message,
		ErrorMessage: errMessage,
		AfterData:    after,
	}, "WalletHandler")
}

func writeWalletError(ctx *gin.Context, logPrefix string, logID uuid.UUID, err error) {
	logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("%s; Error: %+v", logPrefix, err))

	if errors.Is(err, gorm.ErrRecordNotFound) || servicewallet.IsNotFound(err) {
		res := response.Response(http.StatusNotFound, messages.MsgNotFound, logID, nil)
		res.Error = response.Errors{Code: http.StatusNotFound, Message: "wallet resource not found"}
		ctx.JSON(http.StatusNotFound, res)
		return
	}
	if servicewallet.IsPublicError(err) {
		res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logID, nil)
		res.Error = response.Errors{Code: http.StatusBadRequest, Message: walletPublicErrorMessage(err)}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	ctx.JSON(http.StatusInternalServerError, response.InternalServerError(logID))
}

func walletPublicErrorMessage(err error) string {
	switch {
	case errors.Is(err, domainwallet.ErrInsufficientMainBalance):
		return "insufficient main provider balance; top up H2H main balance first, then approve this pending deposit again"
	case errors.Is(err, domainwallet.ErrMainBalanceUnavailable):
		return "main provider balance unavailable; check H2H credentials/connectivity before approving deposit"
	default:
		return err.Error()
	}
}

func writeBadRequest(ctx *gin.Context, logID uuid.UUID, message string) {
	res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logID, nil)
	res.Error = response.Errors{Code: http.StatusBadRequest, Message: message}
	ctx.JSON(http.StatusBadRequest, res)
}

func writeBindError(ctx *gin.Context, logID uuid.UUID, req interface{}, err error) {
	logger.WriteLogWithContext(ctx, logger.LogLevelError, fmt.Sprintf("[WalletHandler][BindJSON]; Error: %+v", err))
	res := response.Response(http.StatusBadRequest, messages.InvalidRequest, logID, nil)
	res.Error = utils.ValidateError(err, reflect.TypeOf(req), "json")
	ctx.JSON(http.StatusBadRequest, res)
}
