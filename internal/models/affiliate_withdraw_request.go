package models

import (
	"time"

	"gorm.io/gorm"
)

// AffiliateWithdrawRequest 推广返利提现申请
type AffiliateWithdrawRequest struct {
	ID                 uint           `gorm:"primarykey" json:"id"`                                // 主键
	AffiliateProfileID uint           `gorm:"not null;index" json:"affiliate_profile_id"`          // 推广用户ID
	Amount             Money          `gorm:"type:decimal(20,2);not null;default:0" json:"amount"` // 申请金额
	Fee                Money          `gorm:"type:decimal(10,2);default:0" json:"fee"`             // 手续费
	ActualAmount       Money          `gorm:"type:decimal(10,2)" json:"actual_amount"`             // 实际到账金额
	Channel            string         `gorm:"type:varchar(50);not null" json:"channel"`            // 提现渠道（alipay-支付宝）
	Account            string         `gorm:"type:varchar(255);not null" json:"account"`           // 支付宝账号
	AlipayName         string         `gorm:"type:varchar(50)" json:"alipay_name"`                 // 真实姓名
	Status             string         `gorm:"type:varchar(32);not null;index" json:"status"`       // 提现状态
	RejectReason       string         `gorm:"type:varchar(255)" json:"reject_reason"`              // 拒绝原因
	BalanceTxnID       *uint          `json:"balance_txn_id,omitempty"`                            // 关联余额账变ID
	ProcessedBy        *uint          `gorm:"index" json:"processed_by,omitempty"`                 // 审核管理员ID
	ProcessedAt        *time.Time     `gorm:"index" json:"processed_at,omitempty"`                 // 审核时间
	CreatedAt          time.Time      `gorm:"index" json:"created_at"`                             // 创建时间
	UpdatedAt          time.Time      `gorm:"index" json:"updated_at"`                             // 更新时间
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`                                      // 软删除时间

	AffiliateProfile AffiliateProfile `gorm:"foreignKey:AffiliateProfileID" json:"affiliate_profile,omitempty"` // 推广用户
	Processor        *Admin           `gorm:"foreignKey:ProcessedBy" json:"processor,omitempty"`                // 审核管理员
}

// TableName 指定表名
func (AffiliateWithdrawRequest) TableName() string {
	return "affiliate_withdraw_requests"
}
