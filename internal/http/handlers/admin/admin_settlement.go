package admin

import (
	"net/http"
	"strconv"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/gin-gonic/gin"
)

// GetAdminWithdrawRequests GET /api/v1/admin/settlement/withdraw-requests
func (h *Handler) GetAdminWithdrawRequests(c *gin.Context) {
	status := c.Query("status")
	page := queryIntDefault(c, "page", 1)
	pageSize := queryIntDefault(c, "page_size", 20)

	requests, total, err := h.SettlementService.GetAdminWithdrawRequests(status, page, pageSize)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, gin.H{
		"items":     requests,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type rejectWithdrawRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// RejectWithdraw POST /api/v1/admin/settlement/withdraw-reject/:id
func (h *Handler) RejectWithdraw(c *gin.Context) {
	id, err := shared.ParseParamUint(c, "id")
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req rejectWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}

	// 获取admin信息
	admin, err := h.AdminRepo.GetByID(adminID)
	if err != nil || admin == nil {
		shared.RespondError(c, response.CodeNotFound, "error.admin_not_found", nil)
		return
	}

	if err := h.SettlementService.AdminRejectWithdraw(id, admin.ID, admin.Username, req.Reason); err != nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}
	response.Success(c, nil)
}

type completeWithdrawRequest struct {
	TransactionID string `json:"transaction_id"`
}

// CompleteWithdraw POST /api/v1/admin/settlement/withdraw-complete/:id
func (h *Handler) CompleteWithdraw(c *gin.Context) {
	id, err := shared.ParseParamUint(c, "id")
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req completeWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}

	// 获取admin信息
	admin, err := h.AdminRepo.GetByID(adminID)
	if err != nil || admin == nil {
		shared.RespondError(c, response.CodeNotFound, "error.admin_not_found", nil)
		return
	}

	if err := h.SettlementService.AdminCompleteWithdraw(id, admin.ID, admin.Username, req.TransactionID); err != nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}
	response.Success(c, nil)
}

// GetWithdrawSettings GET /api/v1/admin/settlement/settings
func (h *Handler) GetAdminWithdrawSettings(c *gin.Context) {
	settings, err := h.SettlementService.GetWithdrawSettings()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, settings)
}

// UpdateWithdrawSettings PUT /api/v1/admin/settlement/settings
func (h *Handler) UpdateAdminWithdrawSettings(c *gin.Context) {
	var req models.AffiliateWithdrawSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	if err := h.SettlementService.UpdateWithdrawSettings(req); err != nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}
	response.Success(c, nil)
}

// queryIntDefault 解析 query 参数为 int，失败时返回默认值
func queryIntDefault(c *gin.Context, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return def
	}
	return v
}
