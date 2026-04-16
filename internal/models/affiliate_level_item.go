package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateLevelItem 推广员等级方案的每个档位（最多3个）
type AffiliateLevelItem struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	SchemeID  uint           `gorm:"not null;index" json:"scheme_id"`
	SortOrder int            `gorm:"not null;default:0" json:"sort_order"`
	Name      string         `gorm:"type:varchar(32);not null" json:"name"`
	Icon      string         `gorm:"type:varchar(16);not null;default:''" json:"icon"`
	Rate      float64        `gorm:"type:decimal(5,2);not null;default:0" json:"rate"`
	IsEntry   bool           `gorm:"not null;default:false" json:"is_entry"`
	Style     string         `gorm:"type:varchar(20);not null;default:''" json:"style"`
	UpgradeConditionType  string  `gorm:"type:varchar(20);not null;default:''" json:"upgrade_condition_type"`
	UpgradePeriodDays     int     `gorm:"not null;default:0" json:"upgrade_period_days"`
	UpgradeTargetAmount   float64 `gorm:"type:decimal(20,2);not null;default:0" json:"upgrade_target_amount"`
	UpgradeTargetOrders   int     `gorm:"not null;default:0" json:"upgrade_target_orders"`
	UpgradeContinuousDays int     `gorm:"not null;default:0" json:"upgrade_continuous_days"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Scheme AffiliateLevelScheme `gorm:"foreignKey:SchemeID" json:"scheme,omitempty"`
}

func (AffiliateLevelItem) TableName() string {
	return "affiliate_level_items"
}
