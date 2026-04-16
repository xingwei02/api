package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateLevelScheme 推广员等级方案主表（每人一套，最多3个等级）
type AffiliateLevelScheme struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"not null;uniqueIndex" json:"user_id"` // 推广员用户ID，每人唯一一套
	MyRate    float64        `gorm:"type:decimal(5,2);not null;default:0" json:"my_rate"`    // 该推广员自身的利润上限（由上级分配）
	EntryRate float64        `gorm:"type:decimal(5,2);not null;default:0" json:"entry_rate"` // 入门档比例（新伙伴默认拿到的比例）
	Version   int            `gorm:"not null;default:1" json:"version"`                      // 版本号，每次保存+1
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Items []AffiliateLevelItem `gorm:"foreignKey:SchemeID" json:"items,omitempty"`
	User  User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AffiliateLevelScheme) TableName() string {
	return "affiliate_level_schemes"
}
