package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateContact 推广员联系资料
type AffiliateContact struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	UserID              uint           `gorm:"not null;uniqueIndex" json:"user_id"`
	Phone               string         `gorm:"type:varchar(32);not null;default:''" json:"phone"`
	QQ                  string         `gorm:"type:varchar(32);not null;default:''" json:"qq"`
	Wechat              string         `gorm:"type:varchar(64);not null;default:''" json:"wechat"`
	OtherContact        string         `gorm:"type:varchar(255);not null;default:''" json:"other_contact"`
	Announcement        string         `gorm:"type:text;not null;default:''" json:"announcement"`
	Notice              string         `gorm:"type:text;not null;default:''" json:"notice"`
	GroupImageURL       string         `gorm:"type:varchar(512);not null;default:''" json:"group_image_url"`
	ParentGroupImageURL string         `gorm:"type:varchar(512);not null;default:''" json:"parent_group_image_url"`
	CreatedAt           time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	User                User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AffiliateContact) TableName() string {
	return "affiliate_contacts"
}
