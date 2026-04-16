package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateDiscount 推广员客户优惠配置（最高5%）
type AffiliateDiscount struct {
	ID                  uint           `gorm:"primarykey" json:"id" `
	UserID              uint           `gorm:"not null;uniqueIndex" json:"user_id" `
	DiscountRate        float64        `gorm:"type:decimal(5,2);not null;default:0" json:"discount_rate" `
	MerchantPageEnabled bool           `gorm:"not null;default:false" json:"merchant_page_enabled" `
	GroupSectionEnabled bool           `gorm:"not null;default:false" json:"group_section_enabled" `
	CreatedAt           time.Time      `gorm:"index" json:"created_at" `
	UpdatedAt           time.Time      `gorm:"index" json:"updated_at" `
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-" `
	User User `gorm:"foreignKey:UserID" json:"user,omitempty" `
}

func (AffiliateDiscount) TableName() string {
	return "affiliate_discounts"
}
