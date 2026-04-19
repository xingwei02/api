package service

import (
	"log"
	"time"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

// AffiliateLevelUpgradeService 伙伴等级自动升级服务
// 每天凌晨3点（北京时间）触发，检查所有伙伴是否达到升级条件
type AffiliateLevelUpgradeService struct{}

// NewAffiliateLevelUpgradeService 创建等级升级服务
func NewAffiliateLevelUpgradeService() *AffiliateLevelUpgradeService {
	return &AffiliateLevelUpgradeService{}
}

// RunDailyUpgradeCheck 执行每日升级检查（由定时任务调用）
// 逻辑：
//  1. 遍历所有有上级的 UserPromotionLevel 记录
//  2. 查上级的 AffiliateLevelScheme，找到比当前档位更高的档位
//  3. 检查伙伴在考核周期内的销售额/订单数是否达到升级条件
//  4. 达标则升级到对应档位，更新 current_level / current_rate / level_item_id
func (s *AffiliateLevelUpgradeService) RunDailyUpgradeCheck() {
	if models.DB == nil {
		return
	}

	// 北京时间今天的起止（用于按日考核）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	// 按周考核：取最近7天
	weekStart := todayStart.AddDate(0, 0, -6)

	// 查所有有上级关系的伙伴
	var levels []models.UserPromotionLevel
	if err := models.DB.Where("parent_user_id > 0").Find(&levels).Error; err != nil {
		log.Printf("[LevelUpgrade] 查询 UserPromotionLevel 失败: %v", err)
		return
	}

	upgraded := 0
	for _, level := range levels {
		if err := s.checkAndUpgrade(level, todayStart, weekStart); err != nil {
			log.Printf("[LevelUpgrade] 用户 %d 升级检查失败: %v", level.UserID, err)
		} else {
			upgraded++
		}
	}
	log.Printf("[LevelUpgrade] 完成，共检查 %d 人，处理 %d 人", len(levels), upgraded)
}

func (s *AffiliateLevelUpgradeService) checkAndUpgrade(
	level models.UserPromotionLevel,
	todayStart, weekStart time.Time,
) error {
	if models.DB == nil {
		return nil
	}

	// 查上级的等级方案（含档位列表，按 sort_order 升序）
	var scheme models.AffiliateLevelScheme
	if err := models.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order asc, id asc")
	}).Where("user_id = ?", level.ParentUserID).First(&scheme).Error; err != nil {
		return nil // 上级没有方案，跳过
	}
	if len(scheme.Items) == 0 {
		return nil
	}

	// 找到比当前档位更高的档位（rate 更高的）
	currentRate := level.CurrentRate
	var targetItem *models.AffiliateLevelItem
	for i := range scheme.Items {
		item := &scheme.Items[i]
		if item.IsEntry {
			continue // 入门档不作为升级目标
		}
		if item.Rate <= currentRate {
			continue // 比当前档位低或相同，跳过
		}
		// 检查升级条件
		if s.meetsUpgradeCondition(level.UserID, item, todayStart, weekStart) {
			// 取 rate 最高的达标档位
			if targetItem == nil || item.Rate > targetItem.Rate {
				targetItem = item
			}
		}
	}

	if targetItem == nil {
		return nil // 没有达标的更高档位
	}

	// 执行升级
	return models.DB.Model(&models.UserPromotionLevel{}).
		Where("user_id = ?", level.UserID).
		Updates(map[string]interface{}{
			"level_item_id": targetItem.ID,
			"current_level": level.CurrentLevel + 1,
			"current_rate":  targetItem.Rate,
			"updated_at":    time.Now(),
		}).Error
}

// meetsUpgradeCondition 检查伙伴是否满足某档位的升级条件
func (s *AffiliateLevelUpgradeService) meetsUpgradeCondition(
	userID uint,
	item *models.AffiliateLevelItem,
	todayStart, weekStart time.Time,
) bool {
	if item == nil || models.DB == nil {
		return false
	}

	// 没有设置升级条件，不能自动升级
	if item.UpgradePeriodDays <= 0 && item.UpgradeTargetAmount <= 0 && item.UpgradeTargetOrders <= 0 {
		return false
	}

	// 确定考核起始时间
	var periodStart time.Time
	if item.UpgradePeriodDays == 7 {
		periodStart = weekStart
	} else {
		periodStart = todayStart
	}

	// 查该伙伴在考核周期内通过自己推广链接产生的订单
	// 关联：orders.affiliate_profile_id → affiliate_profiles.user_id = userID
	type OrderAgg struct {
		TotalAmount float64
		OrderCount  int64
	}
	var agg OrderAgg
	models.DB.Raw(`
		SELECT
			COALESCE(SUM(o.total_amount), 0) AS total_amount,
			COUNT(o.id) AS order_count
		FROM orders o
		INNER JOIN affiliate_profiles ap ON ap.id = o.affiliate_profile_id
		WHERE ap.user_id = ?
		  AND o.status IN ('paid', 'completed')
		  AND o.paid_at >= ?
		  AND o.deleted_at IS NULL
	`, userID, periodStart).Scan(&agg)

	// 按订单数考核
	if item.UpgradeTargetOrders > 0 {
		if int(agg.OrderCount) < item.UpgradeTargetOrders {
			return false
		}
	}

	// 按销售额考核
	if item.UpgradeTargetAmount > 0 {
		if agg.TotalAmount < item.UpgradeTargetAmount {
			return false
		}
	}

	// 连续天数考核（简化：检查 continuous_days 内每天都有订单）
	if item.UpgradeContinuousDays > 1 {
		continuousStart := todayStart.AddDate(0, 0, -(item.UpgradeContinuousDays - 1))
		type DayCount struct {
			Day   string
			Count int64
		}
		var dayCounts []DayCount
		models.DB.Raw(`
			SELECT DATE(o.paid_at) AS day, COUNT(o.id) AS count
			FROM orders o
			INNER JOIN affiliate_profiles ap ON ap.id = o.affiliate_profile_id
			WHERE ap.user_id = ?
			  AND o.status IN ('paid', 'completed')
			  AND o.paid_at >= ?
			  AND o.deleted_at IS NULL
			GROUP BY DATE(o.paid_at)
		`, userID, continuousStart).Scan(&dayCounts)

		if len(dayCounts) < item.UpgradeContinuousDays {
			return false
		}
	}

	return true
}

// StartDailyUpgradeScheduler 启动每日凌晨3点（北京时间）的升级定时任务
func StartDailyUpgradeScheduler(svc *AffiliateLevelUpgradeService) {
	go func() {
		for {
			next := nextBeijing3AM()
			log.Printf("[LevelUpgrade] 下次升级检查时间: %s", next.Format("2006-01-02 15:04:05"))
			time.Sleep(time.Until(next))
			log.Printf("[LevelUpgrade] 开始执行每日升级检查...")
			svc.RunDailyUpgradeCheck()
		}
	}()
}

// nextBeijing3AM 计算下一个北京时间凌晨3点的时刻
func nextBeijing3AM() time.Time {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, loc)
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
