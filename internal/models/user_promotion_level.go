package models

import (
	"time"

	"gorm.io/gorm"
)

// UserPromotionLevel 用户推广等级进度
type UserPromotionLevel struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	UserID          uint           `gorm:"not null;uniqueIndex" json:"user_id"`
	ParentUserID    uint           `gorm:"index" json:"parent_user_id"`
	LevelItemID     uint           `gorm:"index;default:0" json:"level_item_id"`                    // 当前所在档位 ID（关联 affiliate_level_items）
	MaxRate         float64        `gorm:"type:decimal(5,2);not null;default:0" json:"max_rate"`    // 上级分配的利润上限
	CustomRate      float64        `gorm:"type:decimal(5,2);not null;default:-1" json:"custom_rate"` // 上级单独覆盖的比例（-1 表示未覆盖，使用档位默认值）
	CurrentLevel    int            `gorm:"default:1" json:"current_level"`
	CurrentRate     float64        `gorm:"type:decimal(5,2);default:0" json:"current_rate"`
	UpgradeProgress float64        `gorm:"type:decimal(10,2);default:0" json:"upgrade_progress"`
	CycleStart      time.Time      `gorm:"index" json:"cycle_start"`
	CycleEnd        time.Time      `gorm:"index" json:"cycle_end"`
	CreatedAt       time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserPromotionLevel) TableName() string {
	return "user_promotion_levels"
}
