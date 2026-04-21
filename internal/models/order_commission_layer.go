package models

import (
	"time"
)

// OrderCommissionLayer 订单佣金层级表
type OrderCommissionLayer struct {
	ID                 uint      `gorm:"primarykey" json:"id"`
	OrderID            uint      `gorm:"not null;index:idx_order_id" json:"order_id"`
	LayerNum           int       `gorm:"not null" json:"layer_num"`                                                 // 层级序号（1=直推，2=二级，3=三级...）
	UserID             uint      `gorm:"not null;index:idx_user_id" json:"user_id"`                                 // 该层用户ID
	AffiliateProfileID uint      `gorm:"not null;index:idx_affiliate_profile_id" json:"affiliate_profile_id"`       // 该层推广用户ID
	Role               string    `gorm:"type:varchar(32);not null" json:"role"`                                     // 角色（direct=直推，indirect=间推）
	Rate               Money     `gorm:"type:decimal(5,4);not null" json:"rate"`                                    // 分佣比例（0.2000=20%）
	Amount             Money     `gorm:"type:decimal(20,2);not null" json:"amount"`                                 // 分佣金额
	CommissionID       *uint     `gorm:"index:idx_commission_id" json:"commission_id,omitempty"`                    // 关联的佣金记录ID
	CreatedAt          time.Time `gorm:"index:idx_created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
}

// TableName 指定表名
func (OrderCommissionLayer) TableName() string {
	return "order_commission_layers"
}

// 角色常量
const (
	CommissionRoleDirect   = "direct"   // 直推
	CommissionRoleIndirect = "indirect" // 间推
)
