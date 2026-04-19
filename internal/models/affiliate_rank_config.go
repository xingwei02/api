package models

import "time"

// AffiliateRankConfig 封神榜自定义配置
// 全局一条记录，use_custom=true 时前端显示自定义数据，否则显示真实数据
type AffiliateRankConfig struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 全局开关：true=显示自定义数据，false=显示真实数据
	UseCustom bool `gorm:"not null;default:false" json:"use_custom"`

	// 今日销售额榜首
	TopSalesName   string  `gorm:"type:varchar(64)" json:"top_sales_name"`
	TopSalesAmount float64 `gorm:"type:decimal(20,2);default:0" json:"top_sales_amount"`

	// 今日单王（订单数）
	TopOrdersName  string `gorm:"type:varchar(64)" json:"top_orders_name"`
	TopOrdersCount int    `gorm:"default:0" json:"top_orders_count"`

	// 今日开单王（最早出单时间，格式 HH:MM）
	EarliestOrderName string `gorm:"type:varchar(64)" json:"earliest_order_name"`
	EarliestOrderTime string `gorm:"type:varchar(8)" json:"earliest_order_time"`

	// 今日团队王（团队成交额）
	TopTeamName   string  `gorm:"type:varchar(64)" json:"top_team_name"`
	TopTeamAmount float64 `gorm:"type:decimal(20,2);default:0" json:"top_team_amount"`

	// 今日网络王（网络成交额）
	TopNetworkName   string  `gorm:"type:varchar(64)" json:"top_network_name"`
	TopNetworkAmount float64 `gorm:"type:decimal(20,2);default:0" json:"top_network_amount"`

	// 历史闪电王（注册到首单最短分钟数）
	FastestOrderName    string `gorm:"type:varchar(64)" json:"fastest_order_name"`
	FastestOrderMinutes int    `gorm:"default:0" json:"fastest_order_minutes"`
}
