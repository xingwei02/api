package service

import (
	"errors"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

// PromotionService 推广方案服务
type PromotionService struct {
	db *gorm.DB
}

// NewPromotionService 创建推广服务实例
func NewPromotionService(db *gorm.DB) *PromotionService {
	return &PromotionService{db: db}
}

// GetPromotionPlan 获取用户的推广方案
func (s *PromotionService) GetPromotionPlan(userID uint) (*models.PromotionPlan, error) {
	var plan models.PromotionPlan
	if err := s.db.Where("user_id = ? AND status = ?", userID, "active").First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

// CreateOrUpdatePromotionPlan 创建或更新推广方案
func (s *PromotionService) CreateOrUpdatePromotionPlan(plan *models.PromotionPlan) error {
	if plan.UserID == 0 {
		return errors.New("user_id is required")
	}

	var existing models.PromotionPlan
	result := s.db.Where("user_id = ?", plan.UserID).First(&existing)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		plan.Status = "active"
		return s.db.Create(plan).Error
	}

	return s.db.Model(&existing).Updates(plan).Error
}

// InitializeUserLevel 初始化用户等级
func (s *PromotionService) InitializeUserLevel(userID, parentUserID uint) error {
	userLevel := models.UserPromotionLevel{
		UserID:       userID,
		ParentUserID: parentUserID,
		CurrentLevel: 1,
		CurrentRate:  0,
	}
	return s.db.Create(&userLevel).Error
}

// RecordCycleData 记录周期数据
func (s *PromotionService) RecordCycleData(cycleData *models.CycleData) error {
	return s.db.Create(cycleData).Error
}

// GetUserPromotionLevel 获取用户等级
func (s *PromotionService) GetUserPromotionLevel(userID uint) (*models.UserPromotionLevel, error) {
	var level models.UserPromotionLevel
	if err := s.db.Where("user_id = ?", userID).First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &level, nil
}

// CheckUpgrade 检查升级条件
func (s *PromotionService) CheckUpgrade(userID uint) (bool, error) {
	userLevel, err := s.GetUserPromotionLevel(userID)
	if err != nil || userLevel == nil {
		return false, err
	}

	if userLevel.CurrentLevel >= 3 {
		return false, nil
	}

	return true, nil
}

// UpdateUserLevel 更新用户等级
func (s *PromotionService) UpdateUserLevel(userID uint, newLevel int, newRate float64) error {
	return s.db.Model(&models.UserPromotionLevel{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"current_level": newLevel,
			"current_rate":  newRate,
		}).Error
}

// CalculateCommission 计算返利
func (s *PromotionService) CalculateCommission(orderAmount float64, userRate float64) float64 {
	return orderAmount * userRate / 100
}

// GetUserProgress 获取用户进度
func (s *PromotionService) GetUserProgress(userID uint) (map[string]interface{}, error) {
	userLevel, err := s.GetUserPromotionLevel(userID)
	if err != nil {
		return nil, err
	}

	if userLevel == nil {
		return map[string]interface{}{
			"current_level": 0,
			"current_rate":  0,
		}, nil
	}

	return map[string]interface{}{
		"current_level": userLevel.CurrentLevel,
		"current_rate":  userLevel.CurrentRate,
		"progress":      userLevel.UpgradeProgress,
	}, nil
}
