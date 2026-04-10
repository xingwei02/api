package public

import (
	"net/http"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

// UserPromotionProgressResponse 用户推广进度响应
type UserPromotionProgressResponse struct {
	CurrentLevel    int         `json:"current_level"`
	CurrentRate     float64     `json:"current_rate"`
	CycleStart      string      `json:"cycle_start"`
	CycleEnd        string      `json:"cycle_end"`
	ConditionType   string      `json:"condition_type"`
	ConditionValue  float64     `json:"condition_value"`
	CurrentProgress float64     `json:"current_progress"`
	ProgressPercent float64     `json:"progress_percent"`
	OrderCount      int         `json:"order_count"`
}

// GetUserPromotionProgress 获取用户推广进度
func (h *Handler) GetUserPromotionProgress(c *gin.Context) {
	user := shared.MustGetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	promotionSvc := service.NewPromotionService(models.DB)
	progress, err := promotionSvc.GetUserProgress(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "promotion progress not found"})
		return
	}

	resp := UserPromotionProgressResponse{
		CurrentLevel:    progress["current_level"].(int),
		CurrentRate:     progress["current_rate"].(float64),
		CycleStart:      progress["cycle_start"].(string),
		CycleEnd:        progress["cycle_end"].(string),
		ConditionType:   progress["condition_type"].(string),
		ConditionValue:  progress["condition_value"].(float64),
		CurrentProgress: progress["current_progress"].(float64),
		ProgressPercent: progress["progress_percent"].(float64),
		OrderCount:      progress["order_count"].(int),
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UserPromotionHistoryItem 用户推广历史项
type UserPromotionHistoryItem struct {
	Date        string  `json:"date"`
	SalesAmount float64 `json:"sales_amount"`
	OrderCount  int     `json:"order_count"`
}

// GetUserPromotionHistory 获取用户推广历史
func (h *Handler) GetUserPromotionHistory(c *gin.Context) {
	user := shared.MustGetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 获取最近 30 天的数据
	var cycleData []models.CycleData
	if err := models.DB.Where("user_id = ? AND cycle_date >= DATE_SUB(NOW(), INTERVAL 30 DAY)", user.ID).
		Order("cycle_date DESC").
		Find(&cycleData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var items []UserPromotionHistoryItem
	for _, data := range cycleData {
		items = append(items, UserPromotionHistoryItem{
			Date:        data.CycleDate.Format("2006-01-02"),
			SalesAmount: float64(data.SalesAmount),
			OrderCount:  data.OrderCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// UserPromotionCommissionItem 用户佣金项
type UserPromotionCommissionItem struct {
	OrderNo           string  `json:"order_no"`
	CommissionAmount  float64 `json:"commission_amount"`
	CommissionRate    float64 `json:"commission_rate"`
	Status            string  `json:"status"`
	ConfirmAt         string  `json:"confirm_at,omitempty"`
	AvailableAt       string  `json:"available_at,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

// GetUserPromotionCommissions 获取用户佣金列表
func (h *Handler) GetUserPromotionCommissions(c *gin.Context) {
	user := shared.MustGetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 获取用户的推广档案
	var profile models.AffiliateProfile
	if err := models.DB.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "affiliate profile not found"})
		return
	}

	// 获取佣金列表
	var commissions []models.AffiliateCommission
	if err := models.DB.Where("affiliate_profile_id = ?", profile.ID).
		Order("created_at DESC").
		Limit(100).
		Find(&commissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var items []UserPromotionCommissionItem
	for _, comm := range commissions {
		var order models.Order
		if err := models.DB.First(&order, comm.OrderID).Error; err == nil {
			item := UserPromotionCommissionItem{
				OrderNo:          order.OrderNo,
				CommissionAmount: float64(comm.CommissionAmount),
				CommissionRate:   float64(comm.RatePercent),
				Status:           comm.Status,
				CreatedAt:        comm.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			if comm.ConfirmAt != nil {
				item.ConfirmAt = comm.ConfirmAt.Format("2006-01-02 15:04:05")
			}
			if comm.AvailableAt != nil {
				item.AvailableAt = comm.AvailableAt.Format("2006-01-02 15:04:05")
			}
			items = append(items, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}
