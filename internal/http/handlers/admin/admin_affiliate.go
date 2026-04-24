package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAffiliateSettings 获取推广返利设置
func (h *Handler) GetAffiliateSettings(c *gin.Context) {
	setting, err := h.SettingService.GetAffiliateSetting()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, setting)
}

// UpdateAffiliateSettings 更新推广返利设置
func (h *Handler) UpdateAffiliateSettings(c *gin.Context) {
	var req service.AffiliateSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	setting, err := h.SettingService.UpdateAffiliateSetting(req)
	if err != nil {
		if errors.Is(err, service.ErrAffiliateConfigInvalid) {
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		return
	}
	response.Success(c, setting)
}

// GetRankConfig GET /admin/affiliate/rank-config
// 获取封神榜自定义配置
func (h *Handler) GetRankConfig(c *gin.Context) {
	if h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	cfg, err := h.ZhengyeService.GetRankConfig()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, cfg)
}

// BackfillCommissions POST /admin/affiliate/backfill-commissions
// 补偿缺失的订单分佣数据
func (h *Handler) BackfillCommissions(c *gin.Context) {
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}

	type backfillRequest struct {
		OrderIDs  []uint `json:"order_ids"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		DryRun    bool   `json:"dry_run"`
	}

	var req backfillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	// 方式1：指定订单ID列表
	if len(req.OrderIDs) > 0 {
		result, err := h.AffiliateService.BatchBackfillCommissions(req.OrderIDs)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.Success(c, result)
		return
	}

	// 方式2：按日期范围补偿
	startDate := time.Now().AddDate(0, -1, 0) // 默认最近1个月
	endDate := time.Now()
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = t.AddDate(0, 0, 1)
		}
	}

	result, err := h.AffiliateService.BackfillMissingCommissions(service.BackfillCommissionsInput{
		StartDate: startDate,
		EndDate:   endDate,
		DryRun:    req.DryRun,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

// BackfillSingleOrderCommission POST /admin/affiliate/backfill-order/:id
// 为单个订单补偿分佣
func (h *Handler) BackfillSingleOrderCommission(c *gin.Context) {
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || orderID == 0 {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}

	if err := h.AffiliateService.BackfillCommissionsForOrder(uint(orderID)); err != nil {
		if strings.Contains(err.Error(), "已有分佣记录") {
			response.Success(c, gin.H{"message": err.Error(), "skipped": true})
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "分佣补偿成功", "order_id": orderID})
}

// SaveRankConfig PUT /admin/affiliate/rank-config
// 保存封神榜自定义配置（全局开关 + 每维度自定义名字/数值）
func (h *Handler) SaveRankConfig(c *gin.Context) {
	if h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	var req models.AffiliateRankConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	cfg, err := h.ZhengyeService.SaveRankConfig(req)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}
	response.Success(c, cfg)
}
