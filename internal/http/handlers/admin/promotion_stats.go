package admin

import (
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type adminPromotionSubordinateItem struct {
	ID                uint    `json:"id"`
	Email             string  `json:"email"`
	CurrentLevel      int     `json:"currentLevel"`
	CurrentRate       float64 `json:"currentRate"`
	ProgressPercent   float64 `json:"progressPercent"`
	MonthlySales      float64 `json:"monthlySales"`
	MonthlyOrders     int64   `json:"monthlyOrders"`
	MonthlyCommission float64 `json:"monthlyCommission"`
}

type adminPromotionStatsPayload struct {
	TotalSubordinates   int64   `json:"totalSubordinates"`
	PendingCommission   float64 `json:"pendingCommission"`
	AvailableCommission float64 `json:"availableCommission"`
	WithdrawnCommission float64 `json:"withdrawnCommission"`
}

// ListPromotionSubordinates GET /api/v1/admin/affiliate/promotion/subordinates
func (h *Handler) ListPromotionSubordinates(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}

	var levels []models.UserPromotionLevel
	if err := models.DB.Where("parent_user_id = ?", adminID).Order("id desc").Find(&levels).Error; err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}

	items := make([]adminPromotionSubordinateItem, 0, len(levels))
	for _, level := range levels {
		var user models.User
		if err := models.DB.Select("id", "email", "display_name").First(&user, level.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
			return
		}

		var salesRow struct {
			Total  float64
			Orders int64
		}
		if err := models.DB.Model(&models.Order{}).
			Select("COALESCE(SUM(CAST(final_amount AS REAL)), 0) AS total, COUNT(*) AS orders").
			Where("user_id = ?", level.UserID).
			Scan(&salesRow).Error; err != nil {
			shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
			return
		}

		var commissionRow struct {
			Total float64
		}
		if err := models.DB.Model(&models.AffiliateCommission{}).
			Select("COALESCE(SUM(CAST(commission_amount AS REAL)), 0) AS total").
			Where("affiliate_profile_id IN (?)",
				models.DB.Model(&models.AffiliateProfile{}).Select("id").Where("user_id = ?", level.UserID),
			).
			Where("status <> ?", constants.AffiliateCommissionStatusRejected).
			Scan(&commissionRow).Error; err != nil {
			shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
			return
		}

		progress := level.UpgradeProgress
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}

		items = append(items, adminPromotionSubordinateItem{
			ID:                level.UserID,
			Email:             user.Email,
			CurrentLevel:      level.CurrentLevel,
			CurrentRate:       level.CurrentRate,
			ProgressPercent:   progress,
			MonthlySales:      salesRow.Total,
			MonthlyOrders:     salesRow.Orders,
			MonthlyCommission: commissionRow.Total,
		})
	}

	response.Success(c, items)
}

// GetPromotionStats GET /api/v1/admin/affiliate/promotion/stats
func (h *Handler) GetPromotionStats(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}

	stats := adminPromotionStatsPayload{}

	if err := models.DB.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ?", adminID).Count(&stats.TotalSubordinates).Error; err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}

	profileIDs := models.DB.Model(&models.AffiliateProfile{}).Select("id").Where("user_id IN (?)",
		models.DB.Model(&models.UserPromotionLevel{}).Select("user_id").Where("parent_user_id = ?", adminID),
	)

	if err := models.DB.Model(&models.AffiliateCommission{}).
		Select("COALESCE(SUM(CAST(commission_amount AS REAL)), 0)").
		Where("affiliate_profile_id IN (?) AND status = ?", profileIDs, constants.AffiliateCommissionStatusPendingConfirm).
		Scan(&stats.PendingCommission).Error; err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}

	if err := models.DB.Model(&models.AffiliateCommission{}).
		Select("COALESCE(SUM(CAST(commission_amount AS REAL)), 0)").
		Where("affiliate_profile_id IN (?) AND status = ?", profileIDs, constants.AffiliateCommissionStatusAvailable).
		Scan(&stats.AvailableCommission).Error; err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}

	if err := models.DB.Model(&models.AffiliateCommission{}).
		Select("COALESCE(SUM(CAST(commission_amount AS REAL)), 0)").
		Where("affiliate_profile_id IN (?) AND status = ?", profileIDs, constants.AffiliateCommissionStatusWithdrawn).
		Scan(&stats.WithdrawnCommission).Error; err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}

	response.Success(c, stats)
}