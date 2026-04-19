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

// EnsureTokenMerchant 校验当前用户是否已开通 Token 商身份。
func (s *ZhengyeService) EnsureTokenMerchant(userID uint) error {
	if s == nil || s.db == nil || userID == 0 {
		return ErrTokenMerchantRequired
	}
	var user models.User
	if err := s.db.Select("id", "status", "is_token_merchant").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return err
	}
	if user.Status == "disabled" {
		return ErrUserDisabled
	}
	if !user.IsTokenMerchant {
		return ErrTokenMerchantRequired
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// DTO 定义
// ─────────────────────────────────────────────────────────────────────────────

// ZhengyeDashboardDTO 首页概览数据
type ZhengyeDashboardDTO struct {
	AffiliateCode     string  `json:"affiliate_code"`
	PromotionPath     string  `json:"promotion_path"`
	TotalEarnings     string  `json:"total_earnings"`
	TodayEarnings     string  `json:"today_earnings"`
	PaidCommission    string  `json:"paid_commission"`
	PendingCommission string  `json:"pending_commission"`
	TotalSales        string  `json:"total_sales"`
	TotalPartners     int64   `json:"total_partners"`
	DirectPartners    int64   `json:"direct_partners"`
	ActivePartners    int64   `json:"active_partners"`
	TotalOrders       int64   `json:"total_orders"`
	TodayOrders       int64   `json:"today_orders"`
	MyRate            float64 `json:"my_rate"`
	MaxCommissionRate float64 `json:"max_commission_rate"`
	EntryRate         float64 `json:"entry_rate"`
	UpgradeCondition  string  `json:"upgrade_condition"`
	DiscountRate      float64 `json:"discount_rate"`
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

// ZhengyeStatsTodayDTO 今日统计
type ZhengyeStatsTodayDTO struct {
	TotalSales            string `json:"total_sales"`
	SelfSales             string `json:"self_sales"`
	NetSales              string `json:"net_sales"`
	SelfNetSales          string `json:"self_net_sales"`
	NetSettlement         string `json:"net_settlement"`
	NetSettlementOriginal string `json:"net_settlement_original"`
	NetSettlementRefund   string `json:"net_settlement_refund"`
	NetworkOrders         int64  `json:"network_orders"`
	SelfOrders            int64  `json:"self_orders"`
	NewChannels           int64  `json:"new_channels"`
	RefundAmount          string `json:"refund_amount"`
	SelfRefund            string `json:"self_refund"`
}

// ZhengyeStatsTotalDTO 累计统计
type ZhengyeStatsTotalDTO struct {
	TotalSales            string `json:"total_sales"`
	SelfSales             string `json:"self_sales"`
	NetSales              string `json:"net_sales"`
	SelfNetSales          string `json:"self_net_sales"`
	NetSettlement         string `json:"net_settlement"`
	NetSettlementOriginal string `json:"net_settlement_original"`
	NetSettlementRefund   string `json:"net_settlement_refund"`
	NetworkOrders         int64  `json:"network_orders"`
	SelfOrders            int64  `json:"self_orders"`
	Channels              int64  `json:"channels"`
	RefundAmount          string `json:"refund_amount"`
	SelfRefund            string `json:"self_refund"`
}

// ZhengyeStatsDTO 数据看板
type ZhengyeStatsDTO struct {
	Today             *ZhengyeStatsTodayDTO   `json:"today"`
	Total             *ZhengyeStatsTotalDTO   `json:"total"`
	CommissionRate    float64                 `json:"commission_rate"`
	DiscountRate      float64                 `json:"discount_rate"`
	PaidSettlement    string                  `json:"paid_settlement"`
	PendingSettlement string                  `json:"pending_settlement"`
	Trend             []ZhengyeStatsTrendItem `json:"trend"`
}

// ZhengyeOrderItem 订单记录条目
type ZhengyeOrderItem struct {
	OrderID           uint   `json:"order_id"`
	OrderNo           string `json:"order_no"`
	Channel           string `json:"channel"`
	ProductName       string `json:"product_name"`
	Amount            string `json:"amount"`
	Commission        string `json:"commission"`
	PartnerCommission string `json:"partner_commission"`
	ReferrerCost      string `json:"referrer_cost"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
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
	UserID         uint    `json:"user_id"`
	DisplayName    string  `json:"display_name"`
	Level          int     `json:"level"`
	Rate           float64 `json:"rate"`
	TotalOrders    int64   `json:"total_orders"`
	SelfSales      string  `json:"self_sales"`
	TeamSettlement string  `json:"team_settlement"`
	ChannelCount   int64   `json:"channel_count"`
	IsNew          bool    `json:"is_new"`
	JoinedAt       string  `json:"joined_at"`
}

// ZhengyeTeamFilter 团队列表筛选参数
type ZhengyeTeamFilter struct {
	Page     int
	PageSize int
	Depth    int // 1=直属, 2=二级, 3=三级, 0=全部
	Keyword  string
}

// ZhengyeTeamSummary 团队结构汇总
type ZhengyeTeamSummary struct {
	DirectCount    int64 `json:"direct_count"`
	TotalCount     int64 `json:"total_count"`
	NetworkBuyers  int64 `json:"network_buyers"`
	GraduatedCount int64 `json:"graduated_count"`
}

// ZhengyeTeamDTO 团队结构结果
type ZhengyeTeamDTO struct {
	Summary  ZhengyeTeamSummary  `json:"summary"`
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

// ZhengyeRankDimension 封神榜单个维度
type ZhengyeRankDimension struct {
	Name  string `json:"name"`   // 榜首名字
	Value string `json:"value"`  // 数值（金额/数量/时间）
	Rank  int    `json:"rank"`   // 当前用户排名（0=未上榜）
	MyVal string `json:"my_val"` // 当前用户的值
}

// ZhengyeRankDTO 封神榜结果（多维度）
type ZhengyeRankDTO struct {
	UseCustom     bool                 `json:"use_custom"`
	TopSales      ZhengyeRankDimension `json:"top_sales"`      // 今日销售额榜首
	TopOrders     ZhengyeRankDimension `json:"top_orders"`     // 今日单王
	EarliestOrder ZhengyeRankDimension `json:"earliest_order"` // 今日开单王
	TopTeam       ZhengyeRankDimension `json:"top_team"`       // 今日团队王
	TopNetwork    ZhengyeRankDimension `json:"top_network"`    // 今日网络王
	FastestOrder  ZhengyeRankDimension `json:"fastest_order"`  // 历史闪电王
	Items         []ZhengyeRankItem    `json:"items"`          // 累计佣金排行（TOP20）
}

// ZhengyePartnerItem 我的伙伴条目
type ZhengyePartnerItem struct {
	UserID             uint    `json:"user_id"`
	DisplayName        string  `json:"display_name"`
	Email              string  `json:"email"`
	AffiliateCode      string  `json:"affiliate_code"`
	Level              int     `json:"level"`
	LevelName          string  `json:"level_name"`
	LevelIcon          string  `json:"level_icon"`
	Rate               float64 `json:"rate"`
	MaxRate            float64 `json:"max_rate"`
	TodayDirectSales   string  `json:"today_direct_sales"`
	TotalDirectSales   string  `json:"total_direct_sales"`
	TodayNetworkSales  string  `json:"today_network_sales"`
	TotalNetworkSales  string  `json:"total_network_sales"`
	TotalNetworkOrders int64   `json:"total_network_orders"`
	TodaySettlement    string  `json:"today_settlement"`
	TotalSettlement    string  `json:"total_settlement"`
	GroupVisible       bool    `json:"group_visible"`
	IsNew              bool    `json:"is_new"`
	JoinedAt           string  `json:"joined_at"`
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
	AffiliateCode  string `json:"affiliate_code"`
	PendingAmount  string `json:"pending_amount"`
	SettledAmount  string `json:"settled_amount"`
	SelfSales      string `json:"self_sales"`
	TeamSales      string `json:"team_sales"`
	TotalSales     string `json:"total_sales"`
	SelfOrders     int64  `json:"self_orders"`
	TeamOrders     int64  `json:"team_orders"`
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

	// 已打款佣金（已提现）
	var paidCommission float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND status = ?", profile.ID, "withdrawn").
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&paidCommission)

	// 待打款佣金（可提现）
	var pendingCommission float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND status = ?", profile.ID, "available").
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&pendingCommission)

	// 累计销售额（来自关联订单的 base_amount）
	var totalSales float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ?", profile.ID).
		Select("COALESCE(SUM(base_amount), 0)").Scan(&totalSales)

	// 直属伙伴数（parent_user_id = userID）
	var directPartners int64
	s.db.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ?", userID).Count(&directPartners)

	// 全部伙伴数（同直属，因为 UserPromotionLevel 只记录直属关系）
	totalPartners := directPartners

	var todayOrders int64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND created_at >= ?", profile.ID, today).Count(&todayOrders)

	var totalOrders int64
	s.db.Model(&models.AffiliateCommission{}).Where("affiliate_profile_id = ?", profile.ID).Count(&totalOrders)

	var scheme models.AffiliateLevelScheme
	s.db.Where("user_id = ?", userID).First(&scheme)

	// 升级条件文字
	upgradeCondition := ""
	if scheme.EntryRate > 0 {
		upgradeCondition = fmt.Sprintf("入门档 %.0f%%，继续提升销售额可升级", scheme.EntryRate)
	}

	// 客户折扣率
	var discount models.AffiliateDiscount
	var discountRate float64
	if err := s.db.Where("user_id = ?", userID).First(&discount).Error; err == nil {
		discountRate = discount.DiscountRate
	}

	// 推广链接路径
	promotionPath := fmt.Sprintf("/?aff=%s", profile.AffiliateCode)

	return &ZhengyeDashboardDTO{
		AffiliateCode:     profile.AffiliateCode,
		PromotionPath:     promotionPath,
		TotalEarnings:     zhengyeFormatMoney(totalEarnings),
		TodayEarnings:     zhengyeFormatMoney(todayEarnings),
		PaidCommission:    zhengyeFormatMoney(paidCommission),
		PendingCommission: zhengyeFormatMoney(pendingCommission),
		TotalSales:        zhengyeFormatMoney(totalSales),
		TotalPartners:     totalPartners,
		DirectPartners:    directPartners,
		ActivePartners:    totalPartners,
		TotalOrders:       totalOrders,
		TodayOrders:       todayOrders,
		MyRate:            scheme.MyRate,
		MaxCommissionRate: scheme.MyRate,
		EntryRate:         scheme.EntryRate,
		UpgradeCondition:  upgradeCondition,
		DiscountRate:      discountRate,
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
	todayStart := time.Now().UTC().Truncate(24 * time.Hour)

	// 趋势数据
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
	for _, r := range rows {
		trend = append(trend, ZhengyeStatsTrendItem{
			Date:     r.Day,
			Earnings: zhengyeFormatMoney(r.Earnings),
			Orders:   r.Orders,
		})
	}

	// 今日统计：来自 affiliate_commissions 关联 orders
	var todayNetworkOrders int64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND created_at >= ?", profile.ID, todayStart).
		Count(&todayNetworkOrders)

	var todaySelfOrders int64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND commission_type = ? AND created_at >= ?", profile.ID, "order", todayStart).
		Count(&todaySelfOrders)

	var todayTotalSales float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND created_at >= ?", profile.ID, todayStart).
		Select("COALESCE(SUM(base_amount), 0)").Scan(&todayTotalSales)

	var todayCommission float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND created_at >= ?", profile.ID, todayStart).
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&todayCommission)

	// 今日新增渠道（今日新加入的直属伙伴）
	var todayNewChannels int64
	s.db.Model(&models.UserPromotionLevel{}).
		Where("parent_user_id = ? AND created_at >= ?", userID, todayStart).
		Count(&todayNewChannels)

	// 累计统计
	var totalNetworkOrders int64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ?", profile.ID).
		Count(&totalNetworkOrders)

	var totalSelfOrders int64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND commission_type = ?", profile.ID, "order").
		Count(&totalSelfOrders)

	var totalSalesAmount float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ?", profile.ID).
		Select("COALESCE(SUM(base_amount), 0)").Scan(&totalSalesAmount)

	var totalCommission float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND status IN ?", profile.ID, []string{"available", "withdrawn"}).
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&totalCommission)

	var totalChannels int64
	s.db.Model(&models.UserPromotionLevel{}).
		Where("parent_user_id = ?", userID).
		Count(&totalChannels)

	// 已打款 / 待打款结算
	var paidSettlement float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND status = ?", profile.ID, "withdrawn").
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&paidSettlement)

	var pendingSettlement float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND status = ?", profile.ID, "available").
		Select("COALESCE(SUM(commission_amount), 0)").Scan(&pendingSettlement)

	// 当前佣金比例和折扣率
	var scheme models.AffiliateLevelScheme
	s.db.Where("user_id = ?", userID).First(&scheme)

	var discount models.AffiliateDiscount
	var discountRate float64
	if err := s.db.Where("user_id = ?", userID).First(&discount).Error; err == nil {
		discountRate = discount.DiscountRate
	}

	// 今日自己直销额（commission_type = "order" 表示直接销售）
	var todaySelfSales float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND commission_type = ? AND created_at >= ?", profile.ID, "order", todayStart).
		Select("COALESCE(SUM(base_amount), 0)").Scan(&todaySelfSales)

	// 累计自己直销额
	var totalSelfSales float64
	s.db.Model(&models.AffiliateCommission{}).
		Where("affiliate_profile_id = ? AND commission_type = ?", profile.ID, "order").
		Select("COALESCE(SUM(base_amount), 0)").Scan(&totalSelfSales)

	return &ZhengyeStatsDTO{
		Today: &ZhengyeStatsTodayDTO{
			TotalSales:            zhengyeFormatMoney(todayTotalSales),
			SelfSales:             zhengyeFormatMoney(todaySelfSales),
			NetSales:              zhengyeFormatMoney(todayTotalSales),
			SelfNetSales:          zhengyeFormatMoney(todaySelfSales),
			NetSettlement:         zhengyeFormatMoney(todayCommission),
			NetSettlementOriginal: zhengyeFormatMoney(todayCommission),
			NetSettlementRefund:   "0.00",
			NetworkOrders:         todayNetworkOrders,
			SelfOrders:            todaySelfOrders,
			NewChannels:           todayNewChannels,
			RefundAmount:          "0.00",
			SelfRefund:            "0.00",
		},
		Total: &ZhengyeStatsTotalDTO{
			TotalSales:            zhengyeFormatMoney(totalSalesAmount),
			SelfSales:             zhengyeFormatMoney(totalSelfSales),
			NetSales:              zhengyeFormatMoney(totalSalesAmount),
			SelfNetSales:          zhengyeFormatMoney(totalSelfSales),
			NetSettlement:         zhengyeFormatMoney(totalCommission),
			NetSettlementOriginal: zhengyeFormatMoney(totalCommission),
			NetSettlementRefund:   "0.00",
			NetworkOrders:         totalNetworkOrders,
			SelfOrders:            totalSelfOrders,
			Channels:              totalChannels,
			RefundAmount:          "0.00",
			SelfRefund:            "0.00",
		},
		CommissionRate:    scheme.MyRate,
		DiscountRate:      discountRate,
		PaidSettlement:    zhengyeFormatMoney(paidSettlement),
		PendingSettlement: zhengyeFormatMoney(pendingSettlement),
		Trend:             trend,
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
	q.Preload("Order.Items").Order("created_at desc").
		Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&commissions)

	items := make([]ZhengyeOrderItem, 0, len(commissions))
	for _, c := range commissions {
		orderNo := fmt.Sprintf("ORD%08d", c.OrderID)
		productName := ""
		channel := "我的直销"

		if c.Order.ID > 0 {
			orderNo = c.Order.OrderNo
			// 取第一个商品名称
			if len(c.Order.Items) > 0 {
				titleJSON := c.Order.Items[0].TitleJSON
				if s, ok := titleJSON["zh-CN"].(string); ok && s != "" {
					productName = s
				} else if s, ok := titleJSON["en"].(string); ok && s != "" {
					productName = s
				}
			}
			// 判断渠道：如果 affiliate_code 不是当前用户的，则为伙伴渠道
			if c.Order.AffiliateCode != "" {
				var ownerProfile models.AffiliateProfile
				if err := s.db.Where("user_id = ? AND affiliate_code = ?", userID, c.Order.AffiliateCode).First(&ownerProfile).Error; err != nil {
					channel = "伙伴渠道"
				}
			}
		}

		items = append(items, ZhengyeOrderItem{
			OrderID:           c.OrderID,
			OrderNo:           orderNo,
			Channel:           channel,
			ProductName:       productName,
			Amount:            zhengyeFormatMoney(c.BaseAmount.InexactFloat64()),
			Commission:        zhengyeFormatMoney(c.CommissionAmount.InexactFloat64()),
			PartnerCommission: "0.00",
			ReferrerCost:      "0.00",
			Status:            c.Status,
			CreatedAt:         c.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &ZhengyeOrdersDTO{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetTeam 获取团队结构列表（含 summary 汇总）
func (s *ZhengyeService) GetTeam(userID uint, filter ZhengyeTeamFilter) (*ZhengyeTeamDTO, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	// 直属伙伴数
	var directCount int64
	s.db.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ?", userID).Count(&directCount)

	// 全部伙伴数（同直属，因为 UserPromotionLevel 只记录直属关系）
	totalCount := directCount

	// 网络下单用户数：通过当前用户的 affiliate_profile 关联的订单中的 user_id 去重
	var networkBuyers int64
	var myProfile models.AffiliateProfile
	if err := s.db.Where("user_id = ?", userID).First(&myProfile).Error; err == nil {
		s.db.Model(&models.Order{}).
			Where("affiliate_profile_id = ? AND user_id > 0 AND status IN ?", myProfile.ID, []string{"paid", "completed"}).
			Distinct("user_id").Count(&networkBuyers)
	}

	// 已出师合伙人数：升到当前 Token 商设置的最高档位
	var graduatedCount int64
	var maxLevel int
	var scheme models.AffiliateLevelScheme
	if err := s.db.Where("user_id = ?", userID).First(&scheme).Error; err == nil {
		var maxItem models.AffiliateLevelItem
		if err2 := s.db.Where("scheme_id = ?", scheme.ID).Order("sort_order desc, id desc").First(&maxItem).Error; err2 == nil {
			maxLevel = maxItem.SortOrder
		}
	}
	if maxLevel > 0 {
		s.db.Model(&models.UserPromotionLevel{}).
			Where("parent_user_id = ? AND level_item_id IN (SELECT id FROM affiliate_level_items WHERE scheme_id = ? AND sort_order = ?)", userID, scheme.ID, maxLevel).
			Count(&graduatedCount)
	}

	q := s.db.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ?", userID)
	if filter.Keyword != "" {
		q = q.Joins("LEFT JOIN users ON users.id = user_promotion_levels.user_id").
			Where("users.display_name LIKE ? OR users.email LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	var total int64
	q.Count(&total)

	var levels []models.UserPromotionLevel
	q.Preload("User").Order("created_at desc").
		Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&levels)

	todayStart := time.Now().UTC().Truncate(24 * time.Hour)
	sevenDaysAgo := todayStart.AddDate(0, 0, -7)

	items := make([]ZhengyeTeamMember, 0, len(levels))
	for _, l := range levels {
		displayName := ""
		if l.User.ID > 0 {
			displayName = l.User.DisplayName
		}
		var orderCount int64
		var selfSales, teamSettlement float64
		var channelCount int64
		var memberProfile models.AffiliateProfile
		if err := s.db.Where("user_id = ?", l.UserID).First(&memberProfile).Error; err == nil {
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ?", memberProfile.ID).Count(&orderCount)
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND commission_type = ?", memberProfile.ID, "order").
				Select("COALESCE(SUM(base_amount), 0)").Scan(&selfSales)
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND status IN ?", memberProfile.ID, []string{"available", "withdrawn"}).
				Select("COALESCE(SUM(commission_amount), 0)").Scan(&teamSettlement)
		}
		// 该成员的直属下级数（渠道数）
		s.db.Model(&models.UserPromotionLevel{}).Where("parent_user_id = ?", l.UserID).Count(&channelCount)

		isNew := l.CreatedAt.After(sevenDaysAgo)

		items = append(items, ZhengyeTeamMember{
			UserID:         l.UserID,
			DisplayName:    displayName,
			Level:          l.CurrentLevel,
			Rate:           l.CurrentRate,
			TotalOrders:    orderCount,
			SelfSales:      zhengyeFormatMoney(selfSales),
			TeamSettlement: zhengyeFormatMoney(teamSettlement),
			ChannelCount:   channelCount,
			IsNew:          isNew,
			JoinedAt:       l.CreatedAt.Format("2006-01-02"),
		})
	}

	return &ZhengyeTeamDTO{
		Summary: ZhengyeTeamSummary{
			DirectCount:    directCount,
			TotalCount:     totalCount,
			NetworkBuyers:  networkBuyers,
			GraduatedCount: graduatedCount,
		},
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetRank 获取封神榜（多维度，支持自定义数据）
func (s *ZhengyeService) GetRank(currentUserID uint) (*ZhengyeRankDTO, error) {
	// 读取自定义配置
	var cfg models.AffiliateRankConfig
	s.db.First(&cfg) // 不存在时 cfg 为零值，UseCustom=false

	// 累计佣金排行 TOP20（始终查真实数据）
	type rankRow struct {
		UserID      uint
		DisplayName string
		Earnings    float64
		Orders      int64
	}
	var rows []rankRow
	if err := s.db.Table("affiliate_profiles ap").
		Select("ap.user_id as user_id, COALESCE(u.display_name, '') as display_name, COALESCE(SUM(ac.commission_amount), 0) as earnings, COUNT(ac.id) as orders").
		Joins("LEFT JOIN affiliate_commissions ac ON ac.affiliate_profile_id = ap.id").
		Joins("LEFT JOIN users u ON u.id = ap.user_id").
		Group("ap.user_id, u.display_name").
		Order("earnings DESC, orders DESC, ap.user_id ASC").
		Limit(20).
		Scan(&rows).Error; err != nil {
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

	dto := &ZhengyeRankDTO{UseCustom: cfg.UseCustom, Items: items}

	if cfg.UseCustom {
		// 使用自定义数据
		dto.TopSales = ZhengyeRankDimension{Name: cfg.TopSalesName, Value: zhengyeFormatMoney(cfg.TopSalesAmount)}
		dto.TopOrders = ZhengyeRankDimension{Name: cfg.TopOrdersName, Value: fmt.Sprintf("%d", cfg.TopOrdersCount)}
		dto.EarliestOrder = ZhengyeRankDimension{Name: cfg.EarliestOrderName, Value: cfg.EarliestOrderTime}
		dto.TopTeam = ZhengyeRankDimension{Name: cfg.TopTeamName, Value: zhengyeFormatMoney(cfg.TopTeamAmount)}
		dto.TopNetwork = ZhengyeRankDimension{Name: cfg.TopNetworkName, Value: zhengyeFormatMoney(cfg.TopNetworkAmount)}
		dto.FastestOrder = ZhengyeRankDimension{Name: cfg.FastestOrderName, Value: fmt.Sprintf("%d分钟", cfg.FastestOrderMinutes)}
		return dto, nil
	}

	// 使用真实数据
	todayStart := time.Now().UTC().Truncate(24 * time.Hour)

	// 今日销售额榜首（TOP100 中找当前用户排名）
	type salesRow struct {
		UserID      uint
		DisplayName string
		Sales       float64
	}
	var salesRows []salesRow
	s.db.Table("affiliate_profiles ap").
		Select("ap.user_id, COALESCE(u.display_name,'') as display_name, COALESCE(SUM(ac.base_amount),0) as sales").
		Joins("LEFT JOIN affiliate_commissions ac ON ac.affiliate_profile_id = ap.id AND ac.created_at >= ?", todayStart).
		Joins("LEFT JOIN users u ON u.id = ap.user_id").
		Group("ap.user_id, u.display_name").
		Order("sales DESC").Limit(100).Scan(&salesRows)
	if len(salesRows) > 0 {
		dto.TopSales = ZhengyeRankDimension{Name: salesRows[0].DisplayName, Value: zhengyeFormatMoney(salesRows[0].Sales)}
		for i, r := range salesRows {
			if r.UserID == currentUserID {
				dto.TopSales.Rank = i + 1
				dto.TopSales.MyVal = zhengyeFormatMoney(r.Sales)
				break
			}
		}
		if dto.TopSales.Rank == 0 && len(salesRows) >= 100 {
			dto.TopSales.Rank = -1 // 100名以外
		}
	}

	// 今日单王（订单数）
	type ordersRow struct {
		UserID      uint
		DisplayName string
		Cnt         int64
	}
	var ordersRows []ordersRow
	s.db.Table("affiliate_profiles ap").
		Select("ap.user_id, COALESCE(u.display_name,'') as display_name, COUNT(ac.id) as cnt").
		Joins("LEFT JOIN affiliate_commissions ac ON ac.affiliate_profile_id = ap.id AND ac.created_at >= ?", todayStart).
		Joins("LEFT JOIN users u ON u.id = ap.user_id").
		Group("ap.user_id, u.display_name").
		Order("cnt DESC").Limit(100).Scan(&ordersRows)
	if len(ordersRows) > 0 {
		dto.TopOrders = ZhengyeRankDimension{Name: ordersRows[0].DisplayName, Value: fmt.Sprintf("%d单", ordersRows[0].Cnt)}
		for i, r := range ordersRows {
			if r.UserID == currentUserID {
				dto.TopOrders.Rank = i + 1
				dto.TopOrders.MyVal = fmt.Sprintf("%d单", r.Cnt)
				break
			}
		}
		if dto.TopOrders.Rank == 0 && len(ordersRows) >= 100 {
			dto.TopOrders.Rank = -1
		}
	}

	// 今日开单王（最早出单时间）
	type earliestRow struct {
		UserID      uint
		DisplayName string
		FirstOrder  string
	}
	var earliestRows []earliestRow
	s.db.Table("affiliate_profiles ap").
		Select("ap.user_id, COALESCE(u.display_name,'') as display_name, MIN(ac.created_at) as first_order").
		Joins("JOIN affiliate_commissions ac ON ac.affiliate_profile_id = ap.id AND ac.created_at >= ?", todayStart).
		Joins("LEFT JOIN users u ON u.id = ap.user_id").
		Group("ap.user_id, u.display_name").
		Order("first_order ASC").Limit(100).Scan(&earliestRows)
	if len(earliestRows) > 0 {
		t, _ := time.Parse("2006-01-02T15:04:05Z", earliestRows[0].FirstOrder)
		timeStr := t.Format("15:04")
		dto.EarliestOrder = ZhengyeRankDimension{Name: earliestRows[0].DisplayName, Value: timeStr}
		for i, r := range earliestRows {
			if r.UserID == currentUserID {
				dto.EarliestOrder.Rank = i + 1
				t2, _ := time.Parse("2006-01-02T15:04:05Z", r.FirstOrder)
				dto.EarliestOrder.MyVal = t2.Format("15:04")
				break
			}
		}
	}

	// 今日团队王（直属伙伴今日销售额之和）
	type teamRow struct {
		ParentUserID uint
		DisplayName  string
		TeamSales    float64
	}
	var teamRows []teamRow
	s.db.Table("user_promotion_levels upl").
		Select("upl.parent_user_id, COALESCE(u.display_name,'') as display_name, COALESCE(SUM(ac.base_amount),0) as team_sales").
		Joins("JOIN affiliate_profiles ap ON ap.user_id = upl.user_id").
		Joins("LEFT JOIN affiliate_commissions ac ON ac.affiliate_profile_id = ap.id AND ac.created_at >= ?", todayStart).
		Joins("LEFT JOIN users u ON u.id = upl.parent_user_id").
		Group("upl.parent_user_id, u.display_name").
		Order("team_sales DESC").Limit(100).Scan(&teamRows)
	if len(teamRows) > 0 {
		dto.TopTeam = ZhengyeRankDimension{Name: teamRows[0].DisplayName, Value: zhengyeFormatMoney(teamRows[0].TeamSales)}
		for i, r := range teamRows {
			if r.ParentUserID == currentUserID {
				dto.TopTeam.Rank = i + 1
				dto.TopTeam.MyVal = zhengyeFormatMoney(r.TeamSales)
				break
			}
		}
		if dto.TopTeam.Rank == 0 && len(teamRows) >= 100 {
			dto.TopTeam.Rank = -1
		}
	}

	// 今日网络王（自己今日销售额，简化版：等同于今日销售额）
	dto.TopNetwork = dto.TopSales

	// 历史闪电王（Token 商注册到首单最短分钟数）
	type fastRow struct {
		UserID      uint
		DisplayName string
		Minutes     int64
	}
	var fastRows []fastRow
	s.db.Table("users u").
		Select("u.id as user_id, COALESCE(u.display_name,'') as display_name, CAST(EXTRACT(EPOCH FROM (MIN(o.created_at) - u.created_at))/60 AS BIGINT) as minutes").
		Joins("JOIN orders o ON o.user_id = u.id").
		Joins("JOIN affiliate_profiles ap ON ap.user_id = u.id").
		Where("o.status IN ?", []string{"paid", "completed"}).
		Group("u.id, u.display_name").
		Having("MIN(o.created_at) > u.created_at").
		Order("minutes ASC").Limit(100).Scan(&fastRows)
	if len(fastRows) > 0 {
		dto.FastestOrder = ZhengyeRankDimension{Name: fastRows[0].DisplayName, Value: fmt.Sprintf("%d分钟", fastRows[0].Minutes)}
		for i, r := range fastRows {
			if r.UserID == currentUserID {
				dto.FastestOrder.Rank = i + 1
				dto.FastestOrder.MyVal = fmt.Sprintf("%d分钟", r.Minutes)
				break
			}
		}
		if dto.FastestOrder.Rank == 0 && len(fastRows) >= 100 {
			dto.FastestOrder.Rank = -1
		}
	}

	return dto, nil
}

// GetRankConfig 获取封神榜自定义配置（admin 用）
func (s *ZhengyeService) GetRankConfig() (*models.AffiliateRankConfig, error) {
	var cfg models.AffiliateRankConfig
	if err := s.db.First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.AffiliateRankConfig{}, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// SaveRankConfig 保存封神榜自定义配置（admin 用）
func (s *ZhengyeService) SaveRankConfig(input models.AffiliateRankConfig) (*models.AffiliateRankConfig, error) {
	var cfg models.AffiliateRankConfig
	if err := s.db.First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			cfg = input
			cfg.ID = 0
			if err := s.db.Create(&cfg).Error; err != nil {
				return nil, err
			}
			return &cfg, nil
		}
		return nil, err
	}
	cfg.UseCustom = input.UseCustom
	cfg.TopSalesName = input.TopSalesName
	cfg.TopSalesAmount = input.TopSalesAmount
	cfg.TopOrdersName = input.TopOrdersName
	cfg.TopOrdersCount = input.TopOrdersCount
	cfg.EarliestOrderName = input.EarliestOrderName
	cfg.EarliestOrderTime = input.EarliestOrderTime
	cfg.TopTeamName = input.TopTeamName
	cfg.TopTeamAmount = input.TopTeamAmount
	cfg.TopNetworkName = input.TopNetworkName
	cfg.TopNetworkAmount = input.TopNetworkAmount
	cfg.FastestOrderName = input.FastestOrderName
	cfg.FastestOrderMinutes = input.FastestOrderMinutes
	if err := s.db.Save(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
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
		q = q.Joins("LEFT JOIN users ON users.id = user_promotion_levels.user_id").
			Where("users.display_name LIKE ? OR users.email LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	var total int64
	q.Count(&total)

	var levels []models.UserPromotionLevel
	err := q.Preload("User").Order("created_at desc").
		Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize).Find(&levels).Error
	if err != nil {
		return nil, err
	}

	todayStart := time.Now().UTC().Truncate(24 * time.Hour)
	sevenDaysAgo := todayStart.AddDate(0, 0, -7)

	items := make([]ZhengyePartnerItem, 0, len(levels))
	for _, l := range levels {
		// 获取伙伴的 affiliate_profile
		var partnerProfile models.AffiliateProfile
		var todayDirectSales, totalDirectSales, todayCommission, totalCommission float64
		var totalNetworkOrders int64
		var affiliateCode string

		if err := s.db.Where("user_id = ?", l.UserID).First(&partnerProfile).Error; err == nil {
			affiliateCode = partnerProfile.AffiliateCode
			// 今日直销额
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND commission_type = ? AND created_at >= ?", partnerProfile.ID, "order", todayStart).
				Select("COALESCE(SUM(base_amount), 0)").Scan(&todayDirectSales)
			// 累计直销额
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND commission_type = ?", partnerProfile.ID, "order").
				Select("COALESCE(SUM(base_amount), 0)").Scan(&totalDirectSales)
			// 今日佣金（结算）
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND created_at >= ?", partnerProfile.ID, todayStart).
				Select("COALESCE(SUM(commission_amount), 0)").Scan(&todayCommission)
			// 累计佣金（结算）
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND status IN ?", partnerProfile.ID, []string{"available", "withdrawn"}).
				Select("COALESCE(SUM(commission_amount), 0)").Scan(&totalCommission)
			// 网络总订单
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ?", partnerProfile.ID).
				Count(&totalNetworkOrders)
		}

		// 档位名称和图标
		levelName := fmt.Sprintf("等级%d", l.CurrentLevel)
		levelIcon := "🏅"
		if l.LevelItemID > 0 {
			var levelItem models.AffiliateLevelItem
			if err := s.db.First(&levelItem, l.LevelItemID).Error; err == nil {
				levelName = levelItem.Name
				levelIcon = levelItem.Icon
			}
		}

		// 是否新伙伴（7天内加入）
		isNew := l.CreatedAt.After(sevenDaysAgo)

		items = append(items, ZhengyePartnerItem{
			UserID:             l.UserID,
			DisplayName:        l.User.DisplayName,
			Email:              l.User.Email,
			AffiliateCode:      affiliateCode,
			Level:              l.CurrentLevel,
			LevelName:          levelName,
			LevelIcon:          levelIcon,
			Rate:               l.CurrentRate,
			MaxRate:            l.MaxRate,
			TodayDirectSales:   zhengyeFormatMoney(todayDirectSales),
			TotalDirectSales:   zhengyeFormatMoney(totalDirectSales),
			TodayNetworkSales:  zhengyeFormatMoney(todayDirectSales),
			TotalNetworkSales:  zhengyeFormatMoney(totalDirectSales),
			TotalNetworkOrders: totalNetworkOrders,
			TodaySettlement:    zhengyeFormatMoney(todayCommission),
			TotalSettlement:    zhengyeFormatMoney(totalCommission),
			GroupVisible:       true,
			IsNew:              isNew,
			JoinedAt:           l.CreatedAt.Format("2006-01-02"),
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

		// 补充 self_sales/team_sales
		var selfSales, teamSales float64
		if profile.ID > 0 {
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND commission_type = ?", profile.ID, "order").
				Select("COALESCE(SUM(base_amount), 0)").Scan(&selfSales)
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ?", profile.ID).
				Select("COALESCE(SUM(base_amount), 0)").Scan(&teamSales)
		}
		var selfOrders, teamOrders int64
		if profile.ID > 0 {
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ? AND commission_type = ?", profile.ID, "order").Count(&selfOrders)
			s.db.Model(&models.AffiliateCommission{}).
				Where("affiliate_profile_id = ?", profile.ID).Count(&teamOrders)
		}

		items = append(items, ZhengyeSettlementItem{
			UserID:         p.UserID,
			DisplayName:    p.DisplayName,
			AffiliateCode:  p.AffiliateCode,
			PendingAmount:  zhengyeFormatMoney(pending),
			SettledAmount:  zhengyeFormatMoney(settled),
			SelfSales:      zhengyeFormatMoney(selfSales),
			TeamSales:      zhengyeFormatMoney(teamSales),
			TotalSales:     zhengyeFormatMoney(teamSales),
			SelfOrders:     selfOrders,
			TeamOrders:     teamOrders,
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
