package models

import (
	"time"
)

// AffiliateWithdrawSettings 推广佣金提现规则配置
type AffiliateWithdrawSettings struct {
	ID              uint      `gorm:"primarykey" json:"id"`                                         // 主键
	MinAmount       Money     `gorm:"type:decimal(10,2);default:100.00;not null" json:"min_amount"` // 最低提现金额
	IntervalDays    int       `gorm:"default:7;not null" json:"interval_days"`                      // 提现间隔天数
	FeeRate         float64   `gorm:"type:decimal(5,4);default:0.0000;not null" json:"fee_rate"`    // 手续费率（0.01=1%）
	RequireRealname bool      `gorm:"default:true;not null" json:"require_realname"`                // 是否要求实名
	Enabled         bool      `gorm:"default:true;not null" json:"enabled"`                         // 是否开放提现
	CreatedAt       time.Time `json:"created_at"`                                                   // 创建时间
	UpdatedAt       time.Time `json:"updated_at"`                                                   // 更新时间
}

// TableName 指定表名
func (AffiliateWithdrawSettings) TableName() string {
	return "affiliate_withdraw_settings"
}
