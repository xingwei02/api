package public

import (
	"net/http"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

// GetUserPromotionProgress 获取用户推广进度
func (h *Handler) GetUserPromotionProgress(c *gin.Context) {
	userID, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	promotionSvc := service.NewPromotionService(models.DB)
	progress, err := promotionSvc.GetUserProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": progress})
}

// GetUserPromotionLevel 获取用户推广等级
func (h *Handler) GetUserPromotionLevel(c *gin.Context) {
	userID, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	promotionSvc := service.NewPromotionService(models.DB)
	level, err := promotionSvc.GetUserPromotionLevel(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if level == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "promotion level not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": level})
}
