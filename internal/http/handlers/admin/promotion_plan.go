package admin

import (
	"net/http"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// CreateOrUpdatePromotionPlanRequest 创建/更新推广方案请求
type CreateOrUpdatePromotionPlanRequest struct {
	Level1Name      string          `json:"level_1_name" binding:"required"`
	Level1Rate      decimal.Decimal `json:"level_1_rate" binding:"required,gt=0"`
	Level1CondType  string          `json:"level_1_cond_type" binding:"required,oneof=amount count"`
	Level1CondValue decimal.Decimal `json:"level_1_cond_value" binding:"required,gt=0"`
	Level1CondDays  int             `json:"level_1_cond_days" binding:"required,gt=0"`

	Level2Name      string          `json:"level_2_name"`
	Level2Rate      decimal.Decimal `json:"level_2_rate"`
	Level2CondType  string          `json:"level_2_cond_type" binding:"oneof=amount count"`
	Level2CondValue decimal.Decimal `json:"level_2_cond_value"`
	Level2CondDays  int             `json:"level_2_cond_days"`

	Level3Name      string          `json:"level_3_name"`
	Level3Rate      decimal.Decimal `json:"level_3_rate"`
	Level3CondType  string          `json:"level_3_cond_type" binding:"oneof=amount count"`
	Level3CondValue decimal.Decimal `json:"level_3_cond_value"`
	Level3CondDays  int             `json:"level_3_cond_days"`
}

// PromotionPlanResponse 推广方案响应
type PromotionPlanResponse struct {
	ID              uint            `json:"id"`
	UserID          uint            `json:"user_id"`
	Level1Name      string          `json:"level_1_name"`
	Level1Rate      decimal.Decimal `json:"level_1_rate"`
	Level1CondType  string          `json:"level_1_cond_type"`
	Level1CondValue decimal.Decimal `json:"level_1_cond_value"`
	Level1CondDays  int             `json:"level_1_cond_days"`
	Level2Name      string          `json:"level_2_name"`
	Level2Rate      decimal.Decimal `json:"level_2_rate"`
	Level2CondType  string          `json:"level_2_cond_type"`
	Level2CondValue decimal.Decimal `json:"level_2_cond_value"`
	Level2CondDays  int             `json:"level_2_cond_days"`
	Level3Name      string          `json:"level_3_name"`
	Level3Rate      decimal.Decimal `json:"level_3_rate"`
	Level3CondType  string          `json:"level_3_cond_type"`
	Level3CondValue decimal.Decimal `json:"level_3_cond_value"`
	Level3CondDays  int             `json:"level_3_cond_days"`
	Status          string          `json:"status"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

// CreateOrUpdatePromotionPlan 创建或更新推广方案
func (h *Handler) CreateOrUpdatePromotionPlan(c *gin.Context) {
	user := shared.MustGetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateOrUpdatePromotionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promotionSvc := service.NewPromotionService(models.DB)

	plan := &models.PromotionPlan{
		UserID:          user.ID,
		Level1Name:      req.Level1Name,
		Level1Rate:      models.Money(req.Level1Rate.InexactFloat64()),
		Level1CondType:  req.Level1CondType,
		Level1CondValue: models.Money(req.Level1CondValue.InexactFloat64()),
		Level1CondDays:  req.Level1CondDays,
		Level2Name:      req.Level2Name,
		Level2Rate:      models.Money(req.Level2Rate.InexactFloat64()),
		Level2CondType:  req.Level2CondType,
		Level2CondValue: models.Money(req.Level2CondValue.InexactFloat64()),
		Level2CondDays:  req.Level2CondDays,
		Level3Name:      req.Level3Name,
		Level3Rate:      models.Money(req.Level3Rate.InexactFloat64()),
		Level3CondType:  req.Level3CondType,
		Level3CondValue: models.Money(req.Level3CondValue.InexactFloat64()),
		Level3CondDays:  req.Level3CondDays,
	}

	if err := promotionSvc.CreateOrUpdatePromotionPlan(plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := PromotionPlanResponse{
		ID:              plan.ID,
		UserID:          plan.UserID,
		Level1Name:      plan.Level1Name,
		Level1Rate:      decimal.NewFromFloat(float64(plan.Level1Rate)),
		Level1CondType:  plan.Level1CondType,
		Level1CondValue: decimal.NewFromFloat(float64(plan.Level1CondValue)),
		Level1CondDays:  plan.Level1CondDays,
		Level2Name:      plan.Level2Name,
		Level2Rate:      decimal.NewFromFloat(float64(plan.Level2Rate)),
		Level2CondType:  plan.Level2CondType,
		Level2CondValue: decimal.NewFromFloat(float64(plan.Level2CondValue)),
		Level2CondDays:  plan.Level2CondDays,
		Level3Name:      plan.Level3Name,
		Level3Rate:      decimal.NewFromFloat(float64(plan.Level3Rate)),
		Level3CondType:  plan.Level3CondType,
		Level3CondValue: decimal.NewFromFloat(float64(plan.Level3CondValue)),
		Level3CondDays:  plan.Level3CondDays,
		Status:          plan.Status,
		CreatedAt:       plan.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       plan.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GetPromotionPlan 获取推广方案
func (h *Handler) GetPromotionPlan(c *gin.Context) {
	user := shared.MustGetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	promotionSvc := service.NewPromotionService(models.DB)
	plan, err := promotionSvc.GetPromotionPlan(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if plan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "promotion plan not found"})
		return
	}

	resp := PromotionPlanResponse{
		ID:              plan.ID,
		UserID:          plan.UserID,
		Level1Name:      plan.Level1Name,
		Level1Rate:      decimal.NewFromFloat(float64(plan.Level1Rate)),
		Level1CondType:  plan.Level1CondType,
		Level1CondValue: decimal.NewFromFloat(float64(plan.Level1CondValue)),
		Level1CondDays:  plan.Level1CondDays,
		Level2Name:      plan.Level2Name,
		Level2Rate:      decimal.NewFromFloat(float64(plan.Level2Rate)),
		Level2CondType:  plan.Level2CondType,
		Level2CondValue: decimal.NewFromFloat(float64(plan.Level2CondValue)),
		Level2CondDays:  plan.Level2CondDays,
		Level3Name:      plan.Level3Name,
		Level3Rate:      decimal.NewFromFloat(float64(plan.Level3Rate)),
		Level3CondType:  plan.Level3CondType,
		Level3CondValue: decimal.NewFromFloat(float64(plan.Level3CondValue)),
		Level3CondDays:  plan.Level3CondDays,
		Status:          plan.Status,
		CreatedAt:       plan.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       plan.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}
