package models

import "time"

type AffiliateLevelChangeLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	SchemeID   uint      `gorm:"not null;index" json:"scheme_id"`
	BeforeJSON string    `gorm:"type:text;not null;default:''" json:"before_json"`
	AfterJSON  string    `gorm:"type:text;not null;default:''" json:"after_json"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

func (AffiliateLevelChangeLog) TableName() string { return "affiliate_level_change_logs" }
