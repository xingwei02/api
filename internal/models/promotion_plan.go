package models

import (
	"time"

	"gorm.io/gorm"
)

// PromotionPlan 推广方案配置
type PromotionPlan struct {
	ID uint `gorm:"primarykey" json:"id"`

	// 基本信息
	UserID    uint      `gorm:"not null;uniqueIndex" json:"user_id"`     // 推广人ID
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 一级配置
	Level1Name      string `gorm:"type:varchar(30)" json:"level_1_name"`           // 一级等级名称
	Level1Rate      Money  `gorm:"type:decimal(5,2);not null;default:0" json:"level_1_rate"` // 一级返利比例
	Level1CondType  string `gorm:"type:varchar(20);default:'amount'" json:"level_1_cond_type"` // 升级条件类型：amount/count
	Level1CondValue Money  `gorm:"type:decimal(10,2);default:0" json:"level_1_cond_value"`    // 升级条件值
	Level1CondDays  int    `gorm:"default:1" json:"level_1_cond_days"`             // 连续天数

	// 二级配置
	Level2Name      string `gorm:"type:varchar(30)" json:"level_2_name"`
	Level2Rate      Money  `gorm:"type:decimal(5,2);not null;default:0" json:"level_2_rate"`
	Level2CondType  string `gorm:"type:varchar(20);default:'amount'" json:"level_2_cond_type"`
	Level2CondValue Money  `gorm:"type:decimal(10,2);default:0" json:"level_2_cond_value"`
	Level2CondDays  int    `gorm:"default:1" json:"level_2_cond_days"`

	// 三级配置
	Level3Name      string `gorm:"type:varchar(30)" json:"level_3_name"`
	Level3Rate      Money  `gorm:"type:decimal(5,2);not null;default:0" json:"level_3_rate"`
	Level3CondType  string `gorm:"type:varchar(20);default:'amount'" json:"level_3_cond_type"`
	Level3CondValue Money  `gorm:"type:decimal(10,2);default:0" json:"level_3_cond_value"`
	Level3CondDays  int    `gorm:"default:1" json:"level_3_cond_days"`

	// 状态
	Status string `gorm:"type:varchar(20);default:'active';index" json:"status"` // active/inactive

	// 关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (PromotionPlan) TableName() string {
	return "promotion_plans"
}
