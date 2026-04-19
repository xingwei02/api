package admin

import (
	"errors"

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
