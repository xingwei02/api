package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateSettlement 推广员手动结算记录（打款后不可撤销）
type AffiliateSettlement struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	FromUserID uint           `gorm:"not null;index" json:"from_user_id"`
	ToUserID   uint           `gorm:"not null;index" json:"to_user_id"`
	Amount     Money          `gorm:"type:decimal(20,2);not null;default:0" json:"amount"`
	SettleDate string         `gorm:"type:varchar(10);not null;index" json:"settle_date"`
	Status     string         `gorm:"type:varchar(20);not null;index" json:"status"`
	Remark     string         `gorm:"type:varchar(500);not null" json:"remark"`
	SettledAt  *time.Time     `gorm:"index" json:"settled_at,omitempty"`
	CreatedAt  time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	FromUser User `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser   User `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
}

func (AffiliateSettlement) TableName() string {
	return "affiliate_settlements"
}
