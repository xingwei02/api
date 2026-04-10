package models

import (
	"time"

	"gorm.io/gorm"
)

// CycleData 推广考核周期数据
type CycleData struct {
	ID uint `gorm:"primarykey" json:"id"`

	// 关键字段
	UserID      uint      `gorm:"not null;index:idx_user_cycle;index" json:"user_id"`       // 用户ID
	CycleDate   time.Time `gorm:"not null;index:idx_user_cycle" json:"cycle_date"`          // 考核日期
	CycleType   string    `gorm:"type:varchar(20);default:'daily'" json:"cycle_type"`       // 周期类型：daily/weekly

	// 统计数据
	SalesAmount Money `gorm:"type:decimal(10,2);default:0" json:"sales_amount"` // 销售额
	OrderCount  int   `gorm:"default:0" json:"order_count"`                     // 订单数

	// 时间戳
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (CycleData) TableName() string {
	return "cycle_data"
}
