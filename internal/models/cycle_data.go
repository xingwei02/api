package models

import (
	"time"

	"gorm.io/gorm"
)

// CycleData 推广考核周期数据
type CycleData struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null;index:idx_user_cycle;index" json:"user_id"`
	CycleDate   time.Time      `gorm:"not null;index:idx_user_cycle" json:"cycle_date"`
	CycleType   string         `gorm:"type:varchar(20);default:'daily'" json:"cycle_type"`
	SalesAmount float64        `gorm:"type:decimal(10,2);default:0" json:"sales_amount"`
	OrderCount  int            `gorm:"default:0" json:"order_count"`
	CreatedAt   time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (CycleData) TableName() string {
	return "cycle_data"
}
