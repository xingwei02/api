package services

import (
	"errors"
	"time"

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

	// 验证返利比例
	if err := s.validateRates(plan); err != nil {
		return err
	}

	// 检查是否已存在
	var existing models.PromotionPlan
	result := s.db.Where("user_id = ?", plan.UserID).First(&existing)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 新建
		plan.Status = "active"
		return s.db.Create(plan).Error
	} else if result.Error != nil {
		return result.Error
	}

	// 更新
	plan.ID = existing.ID
	return s.db.Model(&existing).Updates(plan).Error
}

// validateRates 验证返利比例（权限递减）
func (s *PromotionService) validateRates(plan *models.PromotionPlan) error {
	// 一级 > 二级 > 三级，且每级相差至少 1%
	if plan.Level1Rate <= 0 {
		return errors.New("level_1_rate must be greater than 0")
	}
	if plan.Level2Rate > 0 && plan.Level2Rate >= plan.Level1Rate {
		return errors.New("level_2_rate must be less than level_1_rate")
	}
	if plan.Level3Rate > 0 && plan.Level3Rate >= plan.Level2Rate {
		return errors.New("level_3_rate must be less than level_2_rate")
	}
	return nil
}

// GetUserPromotionLevel 获取用户的推广等级信息
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

// InitializeUserLevel 初始化用户推广等级（新用户加入推广时调用）
func (s *PromotionService) InitializeUserLevel(userID, parentUserID uint) error {
	// 检查是否已初始化
	var existing models.UserPromotionLevel
	result := s.db.Where("user_id = ?", userID).First(&existing)
	if result.Error == nil {
		return nil // 已存在
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// 获取上级的推广方案
	parentPlan, err := s.GetPromotionPlan(parentUserID)
	if err != nil {
		return err
	}
	if parentPlan == nil {
		return errors.New("parent user has no promotion plan")
	}

	// 创建新的用户等级记录
	now := time.Now().UTC()
	level := models.UserPromotionLevel{
		UserID:       userID,
		ParentUserID: parentUserID,
		CurrentLevel: 1,
		CurrentRate:  parentPlan.Level1Rate,
		CycleStart:   now,
		CycleEnd:     now.AddDate(0, 0, parentPlan.Level1CondDays),
	}

	return s.db.Create(&level).Error
}

// RecordCycleData 记录考核周期数据
func (s *PromotionService) RecordCycleData(userID uint, salesAmount models.Money, orderCount int) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// 检查今天是否已有记录
	var existing models.CycleData
	result := s.db.Where("user_id = ? AND DATE(cycle_date) = DATE(?)", userID, today).First(&existing)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 新建
		data := models.CycleData{
			UserID:      userID,
			CycleDate:   today,
			CycleType:   "daily",
			SalesAmount: salesAmount,
			OrderCount:  orderCount,
		}
		return s.db.Create(&data).Error
	} else if result.Error != nil {
		return result.Error
	}

	// 更新
	return s.db.Model(&existing).Updates(models.CycleData{
		SalesAmount: existing.SalesAmount + salesAmount,
		OrderCount:  existing.OrderCount + orderCount,
	}).Error
}

// CheckUpgrade 检查用户是否满足升级条件
func (s *PromotionService) CheckUpgrade(userID uint) (bool, error) {
	userLevel, err := s.GetUserPromotionLevel(userID)
	if err != nil {
		return false, err
	}
	if userLevel == nil {
		return false, nil
	}

	parentPlan, err := s.GetPromotionPlan(userLevel.ParentUserID)
	if err != nil {
		return false, err
	}
	if parentPlan == nil {
		return false, nil
	}

	// 检查是否已到达最高等级
	if userLevel.CurrentLevel >= 3 {
		return false, nil
	}

	// 获取当前周期的数据
	cycleData, err := s.getCycleDataForPeriod(userID, userLevel.CycleStart, userLevel.CycleEnd)
	if err != nil {
		return false, err
	}

	// 根据当前等级获取升级条件
	var condType string
	var condValue models.Money
	var nextLevel int
	var nextRate models.Money

	if userLevel.CurrentLevel == 1 {
		condType = parentPlan.Level1CondType
		condValue = parentPlan.Level1CondValue
		nextLevel = 2
		nextRate = parentPlan.Level2Rate
	} else if userLevel.CurrentLevel == 2 {
		condType = parentPlan.Level2CondType
		condValue = parentPlan.Level2CondValue
		nextLevel = 3
		nextRate = parentPlan.Level3Rate
	}

	// 检查是否满足条件
	if condType == "amount" {
		if cycleData.SalesAmount >= condValue {
			return true, nil
		}
	} else if condType == "count" {
		if models.Money(cycleData.OrderCount) >= condValue {
			return true, nil
		}
	}

	return false, nil
}

// UpdateUserLevel 更新用户等级
func (s *PromotionService) UpdateUserLevel(userID uint) error {
	shouldUpgrade, err := s.CheckUpgrade(userID)
	if err != nil {
		return err
	}
	if !shouldUpgrade {
		return nil
	}

	userLevel, err := s.GetUserPromotionLevel(userID)
	if err != nil {
		return err
	}
	if userLevel == nil {
		return errors.New("user level not found")
	}

	parentPlan, err := s.GetPromotionPlan(userLevel.ParentUserID)
	if err != nil {
		return err
	}

	// 确定下一个等级和返利比例
	var nextLevel int
	var nextRate models.Money
	var nextCondDays int

	if userLevel.CurrentLevel == 1 {
		nextLevel = 2
		nextRate = parentPlan.Level2Rate
		nextCondDays = parentPlan.Level2CondDays
	} else if userLevel.CurrentLevel == 2 {
		nextLevel = 3
		nextRate = parentPlan.Level3Rate
		nextCondDays = parentPlan.Level3CondDays
	}

	// 更新等级
	now := time.Now().UTC()
	return s.db.Model(userLevel).Updates(models.UserPromotionLevel{
		CurrentLevel:    nextLevel,
		CurrentRate:     nextRate,
		UpgradeProgress: 0,
		CycleStart:      now,
		CycleEnd:        now.AddDate(0, 0, nextCondDays),
	}).Error
}

// CalculateCommission 计算订单的返利
func (s *PromotionService) CalculateCommission(order *models.Order, affiliateUserID uint) (models.Money, error) {
	if order == nil || order.ID == 0 {
		return 0, errors.New("invalid order")
	}

	userLevel, err := s.GetUserPromotionLevel(affiliateUserID)
	if err != nil {
		return 0, err
	}
	if userLevel == nil {
		return 0, errors.New("user promotion level not found")
	}

	// 返利基数 = 订单总金额
	baseAmount := order.TotalAmount

	// 返利金额 = 基数 × 当前返利比例
	commission := baseAmount * userLevel.CurrentRate / 100

	return commission, nil
}

// getCycleDataForPeriod 获取指定周期内的累计数据
func (s *PromotionService) getCycleDataForPeriod(userID uint, startTime, endTime time.Time) (*models.CycleData, error) {
	var result struct {
		TotalSales int64
		TotalCount int64
	}

	if err := s.db.Model(&models.CycleData{}).
		Where("user_id = ? AND cycle_date >= ? AND cycle_date <= ?", userID, startTime, endTime).
		Select("COALESCE(SUM(sales_amount), 0) as total_sales, COALESCE(SUM(order_count), 0) as total_count").
		Scan(&result).Error; err != nil {
		return nil, err
	}

	return &models.CycleData{
		SalesAmount: models.Money(result.TotalSales),
		OrderCount:  int(result.TotalCount),
	}, nil
}

// GetUserProgress 获取用户的升级进度
func (s *PromotionService) GetUserProgress(userID uint) (map[string]interface{}, error) {
	userLevel, err := s.GetUserPromotionLevel(userID)
	if err != nil {
		return nil, err
	}
	if userLevel == nil {
		return nil, errors.New("user promotion level not found")
	}

	parentPlan, err := s.GetPromotionPlan(userLevel.ParentUserID)
	if err != nil {
		return nil, err
	}

	cycleData, err := s.getCycleDataForPeriod(userID, userLevel.CycleStart, userLevel.CycleEnd)
	if err != nil {
		return nil, err
	}

	// 获取升级条件
	var condType string
	var condValue models.Money

	if userLevel.CurrentLevel == 1 {
		condType = parentPlan.Level1CondType
		condValue = parentPlan.Level1CondValue
	} else if userLevel.CurrentLevel == 2 {
		condType = parentPlan.Level2CondType
		condValue = parentPlan.Level2CondValue
	}

	// 计算进度百分比
	var progress float64
	if condType == "amount" {
		progress = float64(cycleData.SalesAmount) / float64(condValue) * 100
	} else if condType == "count" {
		progress = float64(cycleData.OrderCount) / float64(condValue) * 100
	}

	if progress > 100 {
		progress = 100
	}

	return map[string]interface{}{
		"current_level":     userLevel.CurrentLevel,
		"current_rate":      userLevel.CurrentRate,
		"cycle_start":       userLevel.CycleStart,
		"cycle_end":         userLevel.CycleEnd,
		"condition_type":    condType,
		"condition_value":   condValue,
		"current_progress":  cycleData.SalesAmount,
		"progress_percent":  progress,
		"order_count":       cycleData.OrderCount,
	}, nil
}
