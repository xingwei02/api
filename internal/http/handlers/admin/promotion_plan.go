package admin

import (
	"net/http"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

// CreateOrUpdatePromotionPlanRequest 创建/更新推广方案请求
type CreateOrUpdatePromotionPlanRequest struct {
	Level1Name      string  `json:"level_1_name" binding:"required"`
	Level1Rate      float64 `json:"level_1_rate" binding:"required,gt=0"`
	Level1CondType  string  `json:"level_1_cond_type" binding:"required,oneof=amount count"`
	Level1CondValue float64 `json:"level_1_cond_value" binding:"required,gt=0"`
	Level1CondDays  int     `json:"level_1_cond_days" binding:"required,gt=0"`
	Level2Name      string  `json:"level_2_name"`
	Level2Rate      float64 `json:"level_2_rate"`
	Level2CondType  string  `json:"level_2_cond_type"`
	Level2CondValue float64 `json:"level_2_cond_value"`
	Level2CondDays  int     `json:"level_2_cond_days"`
	Level3Name      string  `json:"level_3_name"`
	Level3Rate      float64 `json:"level_3_rate"`
	Level3CondType  string  `json:"level_3_cond_type"`
	Level3CondValue float64 `json:"level_3_cond_value"`
	Level3CondDays  int     `json:"level_3_cond_days"`
}

// CreateOrUpdatePromotionPlan 创建或更新推广方案
func (h *Handler) CreateOrUpdatePromotionPlan(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}

	var req CreateOrUpdatePromotionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promotionSvc := service.NewPromotionService(models.DB)

	plan := &models.PromotionPlan{
		UserID:          adminID,
		Level1Name:      req.Level1Name,
		Level1Rate:      req.Level1Rate,
		Level1CondType:  req.Level1CondType,
		Level1CondValue: req.Level1CondValue,
		Level1CondDays:  req.Level1CondDays,
		Level2Name:      req.Level2Name,
		Level2Rate:      req.Level2Rate,
		Level2CondType:  req.Level2CondType,
		Level2CondValue: req.Level2CondValue,
		Level2CondDays:  req.Level2CondDays,
		Level3Name:      req.Level3Name,
		Level3Rate:      req.Level3Rate,
		Level3CondType:  req.Level3CondType,
		Level3CondValue: req.Level3CondValue,
		Level3CondDays:  req.Level3CondDays,
	}

	if err := promotionSvc.CreateOrUpdatePromotionPlan(plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plan})
}

// GetPromotionPlan 获取推广方案
func (h *Handler) GetPromotionPlan(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}

	promotionSvc := service.NewPromotionService(models.DB)
	plan, err := promotionSvc.GetPromotionPlan(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if plan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "promotion plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plan})
}
