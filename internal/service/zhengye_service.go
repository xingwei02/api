package service

import (
	"fmt"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ZhengyeService 推广中心聚合服务（多级分销体系）
// 独立于旧版 AffiliateService，不破坏现有提现/佣金逻辑。
type ZhengyeService struct {
	db *gorm.DB
}

// NewZhengyeService 创建推广中心聚合服务
func NewZhengyeService(db *gorm.DB) *ZhengyeService {
	return &ZhengyeService{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// DTO 定义
// ─────────────────────────────────────────────────────────────────────────────

// ZhengyeDashboardDTO 首页概览数据
type ZhengyeDashboardDTO struct {
	TotalEarnings  string  `json:"total_earnings"`
	TodayEarnings  string  `json:"today_earnings"`
	TotalPartners  int64   `json:"total_partners"`
	ActivePartners int64   `json:"active_partners"`
	TotalOrders    int64   `json:"total_orders"`
	TodayOrders    int64   `json:"today_orders"`
	MyRate         float64 `json:"my_rate"`
	EntryRate      float64 `json:"entry_rate"`
}

// ZhengyeStatsPeriod 统计周期
type ZhengyeStatsPeriod string

const (
	StatsPeriod7d   ZhengyeStatsPeriod = "7d"
	StatsPeriod30d  ZhengyeStatsPeriod = "30d"
	StatsPeriod180d ZhengyeStatsPeriod = "180d"
)

// ZhengyeStatsTrendItem 趋势数据点
type ZhengyeStatsTrendItem struct {
	Date     string `json:"date"`
	Earnings string `json:"earnings"`
	Orders   int64  `json:"orders"`
}

// ZhengyeStatsDTO 数据看板
type ZhengyeStatsDTO struct {
	TotalEarnings string                  `json:"total_earnings"`
	TotalOrders   int64                   `json:"total_orders"`
	Trend         []ZhengyeStatsTrendItem `json:"trend"`
}

// ZhengyeOrderItem 订单记录条目
type ZhengyeOrderItem struct {
	OrderID    uint   `json:"order_id"`
	Amount     string `json:"amount"`
	Commission string `json:"commission"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// ZhengyeOrdersFilter 订单列表筛选参数
type ZhengyeOrdersFilter struct {
	Page     int
	PageSize int
	Status   string
	DateFrom string
	DateTo   string
}

// ZhengyeOrdersDTO 订单列表结果
type ZhengyeOrdersDTO struct {
	Items    []ZhengyeOrderItem `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// ZhengyeTeamMember 团队成员
type ZhengyeTeamMember struct {
	UserID      uint    `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Level       int     `json:"level"`
	Rate        float64 `json:"rate"`
	TotalOrders int64   `json:"total_orders"`
	JoinedAt    string  `json:"joined_at"`
}

// ZhengyeTeamFilter 团队列表筛选参数
type ZhengyeTeamFilter struct {
	Page     int
	PageSize int
	Depth    int // 1=直属, 2=二级, 3=三级, 0=全部
}

// ZhengyeTeamDTO 团队结构结果
type ZhengyeTeamDTO struct {
	Items    []ZhengyeTeamMember `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// ZhengyeRankItem 封神榜条目
type ZhengyeRankItem struct {
	UserID      uint   `json:"user_id"`
	DisplayName string `json:"display_name"`
	Earnings    string `json:"earnings"`
	Orders      int64  `json:"orders"`
}

// ZhengyeRankDTO 封神榜结果
type ZhengyeRankDTO struct {
	Items []ZhengyeRankItem `json:"items"`
}

// ZhengyePartnerItem 我的伙伴条目
type ZhengyePartnerItem struct {
	UserID      uint    `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Level       int     `json:"level"`
	Rate        float64 `json:"rate"`
	MaxRate     float64 `json:"max_rate"`
	JoinedAt    string  `json:"joined_at"`
}

// ZhengyePartnersFilter 我的伙伴筛选参数
type ZhengyePartnersFilter struct {
	Keyword  string
	Page     int
	PageSize int
}

// ZhengyePartnersDTO 我的伙伴结果
type ZhengyePartnersDTO struct {
	Items    []ZhengyePartnerItem `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// ZhengyeSettlementItem 伙伴结算条目
type ZhengyeSettlementItem struct {
	UserID         uint   `json:"user_id"`
	DisplayName    string `json:"display_name"`
	PendingAmount  string `json:"pending_amount"`
	SettledAmount  string `json:"settled_amount"`
	LastSettlement string `json:"last_settlement"`
	SettleDate     string `json:"settle_date,omitempty"`
}

// ZhengyeSettlementFilter 伙伴结算筛选参数
type ZhengyeSettlementFilter struct {
	Date     string
	Keyword  string
	Page     int
	PageSize int
}

// ZhengyeSettlementDTO 伙伴结算结果
type ZhengyeSettlementDTO struct {
	Items    []ZhengyeSettlementItem `json:"items"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// ZhengyeLevelUpgradeConditionDTO 档位升级条件
type ZhengyeLevelUpgradeConditionDTO struct {
	Days        int     `json:"days,omitempty"`
	DailyAmount float64 `json:"daily_amount,omitempty"`
	Orders      int     `json:"orders,omitempty"`
}

// ZhengyeLevelItemDTO 档位配置条目
type ZhengyeLevelItemDTO struct {
	ID               uint                             `json:"id"`
	Name             string                           `json:"name"`
	Icon             string                           `json:"icon"`
	Rate             float64                          `json:"rate"`
	MemberCount      int64                            `json:"member_count"`
	IsEntry          bool                             `json:"is_entry"`
	UpgradeCondition *ZhengyeLevelUpgradeConditionDTO `json:"upgrade_condition"`
	Style            string                           `json:"style"`
}

// ZhengyeLevelTeamMemberDTO 等级下团队成员
type ZhengyeLevelTeamMemberDTO struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

// ZhengyeLevelTeamGroupDTO 按档位聚合的团队成员
type ZhengyeLevelTeamGroupDTO struct {
	LevelID   uint                        `json:"level_id"`
	LevelName string                      `json:"level_name"`
	Rate      float64                     `json:"rate"`
	Members   []ZhengyeLevelTeamMemberDTO `json:"members"`
}

// ZhengyeLevelsDTO 等级返佣页数据
type ZhengyeLevelsDTO struct {
	MyRate      float64                    `json:"my_rate"`
	EntryRate   float64                    `json:"entry_rate"`
	Levels      []ZhengyeLevelItemDTO      `json:"levels"`
	TeamByLevel []ZhengyeLevelTeamGroupDTO `json:"team_by_level"`
}

// SaveZhengyeLevelsInput 保存等级返佣输入
type SaveZhengyeLevelsInput struct {
	MyRate    float64                     `json:"my_rate"`
	EntryRate float64                     `json:"entry_rate"`
	Levels    []SaveZhengyeLevelItemInput `json:"levels"`
}

// SaveZhengyeLevelItemInput 保存单个档位输入
type SaveZhengyeLevelItemInput struct {
	ID               uint                             `json:"id"`
	Name             string                           `json:"name"`
	Icon             string                           `json:"icon"`
	Rate             float64                          `json:"rate"`
	IsEntry          bool                             `json:"is_entry"`
	UpgradeCondition *ZhengyeLevelUpgradeConditionDTO `json:"upgrade_condition"`
	Style            string                           `json:"style"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Service 方法
// ─────────────────────────────────────────────────────────────────────────────

// GetLevels 获取伙伴等级返佣配置
func (s *ZhengyeService) GetLevels(userID uint) (*ZhengyeLevelsDTO, error) {
	var scheme models.AffiliateLevelScheme
	err := s.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order asc, id asc")
	}).Where("user_id = ?", userID).First(&scheme).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ZhengyeLevelsDTO{
				MyRate:      0,
				EntryRate:   0,
				Levels:      []ZhengyeLevelItemDTO{},
				TeamByLevel: []ZhengyeLevelTeamGroupDTO{},
			}, nil
		}
		return nil, err
	}

	levels := make([]ZhengyeLevelItemDTO, 0, len(scheme.Items))
	teamByLevel := make([]ZhengyeLevelTeamGroupDTO, 0, len(scheme.Items))
	for _, item := range scheme.Items {
		var memberCount int64
		s.db.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ? AND level_item_id = ?", userID, item.ID).Count(&memberCount)

		var memberLevels []models.UserPromotionLevel
		s.db.Preload("User").Where("parent_user_id = ? AND level_item_id = ?", userID, item.ID).Order("created_at desc").Find(&memberLevels)

		members := make([]ZhengyeLevelTeamMemberDTO, 0, len(memberLevels))
		for _, member := range memberLevels {
			members = append(members, ZhengyeLevelTeamMemberDTO{
				ID:     member.UserID,
				Code:   fmt.Sprintf("U%04d", member.UserID),
				Email:  member.User.Email,
				Avatar: "",
			})
		}

		var upgradeCondition *ZhengyeLevelUpgradeConditionDTO
		if item.UpgradePeriodDays > 0 || item.UpgradeTargetAmount > 0 || item.UpgradeTargetOrders > 0 || item.UpgradeContinuousDays > 0 {
			upgradeCondition = &ZhengyeLevelUpgradeConditionDTO{
				Days:        maxInt(item.UpgradeContinuousDays, item.UpgradePeriodDays),
				DailyAmount: item.UpgradeTargetAmount,
				Orders:      item.UpgradeTargetOrders,
			}
		}

		levels = append(levels, ZhengyeLevelItemDTO{
			ID:               item.ID,
			Name:             item.Name,
			Icon:             item.Icon,
			Rate:             item.Rate,
			MemberCount:      memberCount,
			IsEntry:          item.IsEntry,
			UpgradeCondition: upgradeCondition,
			Style:            item.Style,
		})

		teamByLevel = append(teamByLevel, ZhengyeLevelTeamGroupDTO{
			LevelID:   item.ID,
			LevelName: item.Name,
			Rate:      item.Rate,
			Members:   members,
		})
	}

	return &ZhengyeLevelsDTO{
		MyRate:      scheme.MyRate,
		EntryRate:   scheme.EntryRate,
		Levels:      levels,
		TeamByLevel: teamByLevel,
	}, nil
}

// SaveLevels 整体保存伙伴等级返佣配置
func (s *ZhengyeService) SaveLevels(userID uint, input SaveZhengyeLevelsInput) error {
	if len(input.Levels) > 3 {
		return fmt.Errorf("levels cannot exceed 3")
	}
	if input.MyRate < 0 {
		return fmt.Errorf("my_rate cannot be negative")
	}
	if input.EntryRate < 0 {
		return fmt.Errorf("entry_rate cannot be negative")
	}

	entryCount := 0
	for _, item := range input.Levels {
		if item.Name == "" {
			return fmt.Errorf("level name is required")
		}
		if item.Rate < 0 {
			return fmt.Errorf("level rate cannot be negative")
		}
		if input.MyRate > 0 && item.Rate > input.MyRate {
			return fmt.Errorf("level rate cannot exceed my_rate")
		}
		if item.IsEntry {
			entryCount++
		}
		if item.UpgradeCondition != nil {
			if item.UpgradeCondition.Days < 0 || item.UpgradeCondition.DailyAmount < 0 || item.UpgradeCondition.Orders < 0 {
				return fmt.Errorf("upgrade condition cannot be negative")
			}
		}
	}
	if len(input.Levels) > 0 && entryCount != 1 {
		return fmt.Errorf("exactly one entry level is required")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var scheme models.AffiliateLevelScheme
		err := tx.Where("user_id = ?", userID).First(&scheme).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				scheme = models.AffiliateLevelScheme{UserID: userID, Version: 1}
				if err := tx.Create(&scheme).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		scheme.MyRate = input.MyRate
		if entryCount == 1 {
			for _, item := range input.Levels {
				if item.IsEntry {
					scheme.EntryRate = item.Rate
					break
				}
			}
		} else {
			scheme.EntryRate = input.EntryRate
		}
		if scheme.ID > 0 {
			scheme.Version++
		}
		if err := tx.Save(&scheme).Error; err != nil {
			return err
		}

		if err := tx.Where("scheme_id = ?", scheme.ID).Delete(&models.AffiliateLevelItem{}).Error; err != nil {
			return err
		}

		for idx, item := range input.Levels {
			levelItem := models.AffiliateLevelItem{
				SchemeID:  scheme.ID,
				SortOrder: idx + 1,
				Name:      item.Name,
				Icon:      item.Icon,
				Rate:      item.Rate,
				IsEntry:   item.IsEntry,
				Style:     item.Style,
			}
			if item.UpgradeCondition != nil {
				levelItem.UpgradeConditionType = "custom"
				levelItem.UpgradeContinuousDays = item.UpgradeCondition.Days
				levelItem.UpgradePeriodDays = item.UpgradeCondition.Days
				levelItem.UpgradeTargetAmount = item.UpgradeCondition.DailyAmount
				levelItem.UpgradeTargetOrders = item.UpgradeCondition.Orders
			}
			if err := tx.Create(&levelItem).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetDashboard 获取推广中心首页概览数据
func (s *ZhengyeService) GetDashboard(userID uint) (*ZhengyeDashboardDTO, error) {
	var profile models.AffiliateProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	var todayEarnings float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND status = ? AND created_at >= ?", profile.ID, "available", today).
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&todayEarnings)

	var totalEarnings float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND status IN ?", profile.ID, []string{"available", "withdrawn"}).
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&totalEarnings)

	var totalPartners int64
	s.db.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ?", userID).Count(&totalPartners)

	var todayOrders int64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND created_at >= ?", profile.ID, today).Count(&todayOrders)

	var totalOrders int64
	s.db.Model(&models.AffiliateCommission{}).Where("affiliate_profile_id = ?", profile.ID).Count(&totalOrders)

	var scheme models.AffiliateLevelScheme
	s.db.Where("user_id = ?", userID).First(&scheme)

	return &ZhengyeDashboardDTO{
		TotalEarnings:  zhengyeFormatMoney(totalEarnings),
		TodayEarnings:  zhengyeFormatMoney(todayEarnings),
		TotalPartners:  totalPartners,
		ActivePartners: totalPartners,
		TotalOrders:    totalOrders,
		TodayOrders:    todayOrders,
		MyRate:         scheme.MyRate,
		EntryRate:      scheme.EntryRate,
	}, nil
}

// GetStats 获取数据看板（按时间段）
func (s *ZhengyeService) GetStats(userID uint, period ZhengyeStatsPeriod) (*ZhengyeStatsDTO, error) {
	var profile models.AffiliateProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}

	days := 7
	switch period {
	case StatsPeriod30d:
		days = 30
	case StatsPeriod180d:
		days = 180
	}
	startAt := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -days+1)

	type trendRow struct {
		Day      string
		Earnings float64
		Orders   int64
	}
	var rows []trendRow
	s.db.Model(&models.AffiliateCommission{}).
		Select("DATE(created_at) as day, COALESCE(SUM(commission_amount), 0) as earnings, COUNT(*) as orders").
		Where("affiliate_profile_id = ? AND created_at >= ?", profile.ID, startAt).
		Group("DATE(created_at)").Order("day asc").Scan(&rows)

	trend := make([]ZhengyeStatsTrendItem, 0, len(rows))
	var totalEarnings float64
	var totalOrders int64
	for _, r := range rows {
		trend = append(trend, ZhengyeStatsTrendItem{
			Date:     r.Day,
			Earnings: zhengyeFormatMoney(r.Earnings),
			Orders:   r.Orders,
		})
		totalEarnings += r.Earnings
		totalOrders += r.Orders
	}

	return &ZhengyeStatsDTO{
		TotalEarnings: zhengyeFormatMoney(totalEarnings),
		TotalOrders:   totalOrders,
		Trend:         trend,
	}, nil
}

// GetOrders 获取订单记录列表
func (s *ZhengyeService) GetOrders(userID uint, filter ZhengyeOrdersFilter) (*ZhengyeOrdersDTO, error) {
	var profile models.AffiliateProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	q := s.db.Model(&models.AffiliateCommission{}).Where("affiliate_profile_id = ?", profile.ID)
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.DateFrom != "" {
		q = q.Where("created_at >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		q = q.Where("created_at <= ?", filter.DateTo)
	}

	var total int64
	q.Count(&total)

	var commissions []models.AffiliateCommission
	q.Order("created_at desc").Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&commissions)

	items := make([]ZhengyeOrderItem, 0, len(commissions))
	for _, c := range commissions {
		items = append(items, ZhengyeOrderItem{
			OrderID:    c.OrderID,
			Amount:     zhengyeFormatMoney(c.BaseAmount.InexactFloat64()),
			Commission: zhengyeFormatMoney(c.CommissionAmount.InexactFloat64()),
			Status:     c.Status,
			CreatedAt:  c.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &ZhengyeOrdersDTO{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetTeam 获取团队结构列表
func (s *ZhengyeService) GetTeam(userID uint, filter ZhengyeTeamFilter) (*ZhengyeTeamDTO, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	q := s.db.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ?", userID)

	var total int64
	q.Count(&total)

	var levels []models.UserPromotionLevel
	q.Preload("User").Order("created_at desc").
		Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&levels)

	items := make([]ZhengyeTeamMember, 0, len(levels))
	for _, l := range levels {
		displayName := ""
		if l.User.ID > 0 {
			displayName = l.User.DisplayName
		}
		var orderCount int64
		var memberProfile models.AffiliateProfile
		if err := s.db.Where("user_id = ?", l.UserID).First(&memberProfile).Error; err == nil {
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ?", memberProfile.ID).Count(&orderCount)
		}
		items = append(items, ZhengyeTeamMember{
			UserID:      l.UserID,
			DisplayName: displayName,
			Level:       l.CurrentLevel,
			Rate:        l.CurrentRate,
			TotalOrders: orderCount,
			JoinedAt:    l.CreatedAt.Format("2006-01-02"),
		})
	}

	return &ZhengyeTeamDTO{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetRank 获取封神榜（按累计佣金倒序）
func (s *ZhengyeService) GetRank() (*ZhengyeRankDTO, error) {
	type rankRow struct {
		UserID      uint
		DisplayName string
		Earnings    float64
		Orders      int64
	}
	var rows []rankRow
	err := s.db.Table("affiliate_profiles ap").
		Select("ap.user_id as user_id, COALESCE(u.display_name, '') as display_name, COALESCE(SUM(ac.commission_amount), 0) as earnings, COUNT(ac.id) as orders").
		Joins("LEFT JOIN affiliate_commissions ac ON ac.affiliate_profile_id = ap.id").
		Joins("LEFT JOIN users u ON u.id = ap.user_id").
		Group("ap.user_id, u.display_name").
		Order("earnings DESC, orders DESC, ap.user_id ASC").
		Limit(20).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]ZhengyeRankItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, ZhengyeRankItem{
			UserID:      r.UserID,
			DisplayName: r.DisplayName,
			Earnings:    zhengyeFormatMoney(r.Earnings),
			Orders:      r.Orders,
		})
	}
	return &ZhengyeRankDTO{Items: items}, nil
}

// GetPartners 获取我的伙伴（直属）
func (s *ZhengyeService) GetPartners(userID uint, filter ZhengyePartnersFilter) (*ZhengyePartnersDTO, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	q := s.db.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ?", userID)
	if filter.Keyword != "" {
		q = q.Joins("LEFT JOIN users ON users.id = user_promotion_levels.user_id").Where("users.display_name LIKE ? OR users.email LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	var total int64
	q.Count(&total)

	var levels []models.UserPromotionLevel
	err := q.Preload("User").Order("created_at desc").Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&levels).Error
	if err != nil {
		return nil, err
	}

	items := make([]ZhengyePartnerItem, 0, len(levels))
	for _, l := range levels {
		items = append(items, ZhengyePartnerItem{
			UserID:      l.UserID,
			DisplayName: l.User.DisplayName,
			Level:       l.CurrentLevel,
			Rate:        l.CurrentRate,
			MaxRate:     l.MaxRate,
			JoinedAt:    l.CreatedAt.Format("2006-01-02"),
		})
	}

	return &ZhengyePartnersDTO{Items: items, Total: total, Page: filter.Page, PageSize: filter.PageSize}, nil
}

// UpdatePartnerRate 更新直属伙伴自定义比例
func (s *ZhengyeService) UpdatePartnerRate(userID, partnerID uint, rate float64) error {
	var level models.UserPromotionLevel
	if err := s.db.Where("user_id = ? AND parent_user_id = ?", partnerID, userID).First(&level).Error; err != nil {
		return err
	}
	if rate < 0 {
		rate = 0
	}
	if level.MaxRate > 0 && rate > level.MaxRate {
		rate = level.MaxRate
	}
	level.CustomRate = rate
	level.CurrentRate = rate
	return s.db.Save(&level).Error
}

// GetSettlement 获取伙伴结算列表
func (s *ZhengyeService) GetSettlement(userID uint, filter ZhengyeSettlementFilter) (*ZhengyeSettlementDTO, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	if filter.Date == "" {
		filter.Date = time.Now().Format("2006-01-02")
	}

	partnerFilter := ZhengyePartnersFilter{Keyword: filter.Keyword, Page: filter.Page, PageSize: filter.PageSize}
	partners, err := s.GetPartners(userID, partnerFilter)
	if err != nil {
		return nil, err
	}

	items := make([]ZhengyeSettlementItem, 0, len(partners.Items))
	for _, p := range partners.Items {
		var profile models.AffiliateProfile
		var pending float64
		if err := s.db.Where("user_id = ?", p.UserID).First(&profile).Error; err == nil {
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND status = ?", profile.ID, "available").
				Select("COALESCE(SUM(commission_amount), 0)").Scan(&pending)
		}

		var settled float64
		s.db.Model(&models.AffiliateSettlement{}).
			Where("from_user_id = ? AND to_user_id = ? AND status = ?", userID, p.UserID, "paid").
			Select("COALESCE(SUM(amount), 0)").Scan(&settled)

		var last models.AffiliateSettlement
		lastDate := ""
		if err := s.db.Where("from_user_id = ? AND to_user_id = ?", userID, p.UserID).Order("settled_at desc, created_at desc").First(&last).Error; err == nil {
			if last.SettledAt != nil {
				lastDate = last.SettledAt.Format("2006-01-02 15:04:05")
			} else {
				lastDate = last.CreatedAt.Format("2006-01-02 15:04:05")
			}
		}

		items = append(items, ZhengyeSettlementItem{
			UserID:         p.UserID,
			DisplayName:    p.DisplayName,
			PendingAmount:  zhengyeFormatMoney(pending),
			SettledAmount:  zhengyeFormatMoney(settled),
			LastSettlement: lastDate,
			SettleDate:     filter.Date,
		})
	}

	return &ZhengyeSettlementDTO{Items: items, Total: partners.Total, Page: partners.Page, PageSize: partners.PageSize}, nil
}

// PaySettlement 执行手动结算
func (s *ZhengyeService) PaySettlement(userID, partnerID uint, settleDate string) error {
	if settleDate == "" {
		settleDate = time.Now().Format("2006-01-02")
	}
	var level models.UserPromotionLevel
	if err := s.db.Where("user_id = ? AND parent_user_id = ?", partnerID, userID).First(&level).Error; err != nil {
		return err
	}
	var profile models.AffiliateProfile
	if err := s.db.Where("user_id = ?", partnerID).First(&profile).Error; err != nil {
		return err
	}
	var pending float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND status = ?", profile.ID, "available").
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&pending)
	now := time.Now()
	settlement := models.AffiliateSettlement{
		FromUserID: userID,
		ToUserID:   partnerID,
		Amount:     models.NewMoneyFromDecimal(decimal.NewFromFloat(pending)),
		SettleDate: settleDate,
		Status:     "paid",
		Remark:     "manual settlement",
		SettledAt:  &now,
	}
	return s.db.Create(&settlement).Error
}

// zhengyeFormatMoney 将 float64 格式化为两位小数字符串
func zhengyeFormatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
