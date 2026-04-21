package models

import (
	"time"
)

// UserBalance 用户余额表
type UserBalance struct {
	ID                 uint      `gorm:"primarykey" json:"id"`
	UserID             uint      `gorm:"not null;uniqueIndex:uk_user_id" json:"user_id"`
	AffiliateProfileID *uint     `gorm:"index:idx_affiliate_profile_id" json:"affiliate_profile_id,omitempty"` // 推广用户ID（仅Token商有值）
	Balance            float64   `gorm:"type:decimal(20,2);not null;default:0;index" json:"balance"`
	FrozenBalance      float64   `gorm:"type:decimal(20,2);not null;default:0" json:"frozen_balance"`
	TotalIncome        float64   `gorm:"type:decimal(20,2);not null;default:0" json:"total_income"`
	TotalWithdraw      float64   `gorm:"type:decimal(20,2);not null;default:0" json:"total_withdraw"`
	Version            uint      `gorm:"not null;default:1" json:"version"` // 乐观锁
	CreatedAt          time.Time `gorm:"index" json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName 指定表名
func (UserBalance) TableName() string {
	return "user_balances"
}

// UserBalanceLog 用户余额明细账表
type UserBalanceLog struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Type          string    `gorm:"type:varchar(50);not null;index" json:"type"` // commission_transfer/withdraw_apply/withdraw_reject/withdraw_complete/refund/admin_adjust
	Amount        float64   `gorm:"type:decimal(20,2);not null" json:"amount"`   // 正数=入账，负数=出账
	BalanceBefore float64   `gorm:"type:decimal(20,2);not null" json:"balance_before"`
	BalanceAfter  float64   `gorm:"type:decimal(20,2);not null" json:"balance_after"`
	Description   string    `gorm:"type:varchar(500)" json:"description"`
	RelatedType   string    `gorm:"type:varchar(50);index:idx_related" json:"related_type"` // commission/withdraw_request/refund_order
	RelatedID     *uint     `gorm:"index:idx_related" json:"related_id"`
	OperatorID    *uint     `json:"operator_id"`
	OperatorName  string    `gorm:"type:varchar(100)" json:"operator_name"`
	Remark        string    `gorm:"type:varchar(500)" json:"remark"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (UserBalanceLog) TableName() string {
	return "user_balance_logs"
}

// BalanceLogType 余额明细类型常量
const (
	BalanceLogTypeCommissionTransfer = "commission_transfer" // 佣金转入
	BalanceLogTypeWithdrawApply      = "withdraw_apply"      // 提现申请
	BalanceLogTypeWithdrawReject     = "withdraw_reject"     // 提现驳回
	BalanceLogTypeWithdrawComplete   = "withdraw_complete"   // 提现完成
	BalanceLogTypeRefund             = "refund"              // 退款
	BalanceLogTypeAdminAdjust        = "admin_adjust"        // 管理员调整
)
