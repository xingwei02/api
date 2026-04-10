package models

import (
	"time"

	"gorm.io/gorm"
)

// UserPromotionLevel 用户推广等级进度
type UserPromotionLevel struct {
	ID uint `gorm:"primarykey" json:"id"`

	// 基本关系
	UserID       uint `gorm:"not null;uniqueIndex:idx_user_parent;index" json:"user_id"`       // 用户ID
	ParentUserID uint `gorm:"not null;uniqueIndex:idx_user_parent;index" json:"parent_user_id"` // 上级推广人ID

	// 当前等级信息
	CurrentLevel   int   `gorm:"default:1" json:"current_level"`                    // 当前等级（1/2/3）
	CurrentRate    Money `gorm:"type:decimal(5,2);not null;default:0" json:"current_rate"` // 当前返利比例

	// 升级进度
	UpgradeProgress Money      `gorm:"type:decimal(10,2);default:0" json:"upgrade_progress"` // 当前周期进度
	CycleStart      time.Time  `gorm:"index" json:"cycle_start"`                             // 周期开始时间
	CycleEnd        time.Time  `gorm:"index" json:"cycle_end"`                               // 周期结束时间
	ConsecutiveDays int        `gorm:"default:0" json:"consecutive_days"`                    // 连续完成天数

	// 时间戳
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关系
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ParentUser User           `gorm:"foreignKey:ParentUserID;references:ID" json:"parent_user,omitempty"`
	CycleData  []CycleData    `gorm:"foreignKey:UserID" json:"cycle_data,omitempty"`
}

// TableName 指定表名
func (UserPromotionLevel) TableName() string {
	return "user_promotion_levels"
}
