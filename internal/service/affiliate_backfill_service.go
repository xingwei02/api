package service

import (
	"fmt"
	"time"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

// BackfillCommissionsInput 补偿分佣输入参数
type BackfillCommissionsInput struct {
	StartDate time.Time
	EndDate   time.Time
	DryRun    bool // true=仅检查不执行，false=实际执行
}

// BackfillCommissionsResult 补偿分佣结果
type BackfillCommissionsResult struct {
	TotalOrders        int      `json:"total_orders"`        // 总订单数
	FixedOrders        int      `json:"fixed_orders"`        // 已修复订单数
	SuccessCommissions int      `json:"success_commissions"` // 成功生成分佣数
	FailedOrders       []uint   `json:"failed_orders"`       // 失败的订单ID列表
	SkippedOrders      []uint   `json:"skipped_orders"`      // 跳过的订单ID列表
	ErrorMessages      []string `json:"error_messages"`      // 错误信息列表
}

// BackfillMissingCommissions 补偿缺失的订单分佣数据
//
// 功能：
// 1. 查找有上级关系但订单没有 affiliate_profile_id 的情况
// 2. 自动补充订单的推广归因
// 3. 触发分佣计算生成 affiliate_commissions 和 order_commission_layers
//
// 使用场景：
// - 修复历史数据
// - 定期检查和补偿
func (s *AffiliateService) BackfillMissingCommissions(input BackfillCommissionsInput) (*BackfillCommissionsResult, error) {
	if s.repo == nil || s.orderRepo == nil {
		return nil, fmt.Errorf("service not initialized")
	}

	result := &BackfillCommissionsResult{
		FailedOrders:  []uint{},
		SkippedOrders: []uint{},
		ErrorMessages: []string{},
	}

	// 1. 查询需要修复的订单
	type OrderToFix struct {
		OrderID         uint
		UserID          uint
		ParentUserID    uint
		ParentProfileID uint
		ParentCode      string
		Status          string
		PaidAt          *time.Time
	}

	var ordersToFix []OrderToFix
	err := models.DB.Raw(`
		SELECT 
			o.id as order_id,
			o.user_id,
			upl.parent_user_id,
			ap.id as parent_profile_id,
			ap.affiliate_code as parent_code,
			o.status,
			o.paid_at
		FROM orders o
		INNER JOIN user_promotion_levels upl ON upl.user_id = o.user_id
		INNER JOIN affiliate_profiles ap ON ap.user_id = upl.parent_user_id
		LEFT JOIN affiliate_commissions ac ON ac.order_id = o.id
		WHERE o.status IN ('paid', 'completed')
		  AND o.user_id > 0
		  AND o.affiliate_profile_id IS NULL
		  AND o.paid_at BETWEEN ? AND ?
		  AND ap.status = 'active'
		GROUP BY o.id, upl.parent_user_id, ap.id
		HAVING COUNT(ac.id) = 0
		ORDER BY o.created_at ASC
	`, input.StartDate, input.EndDate).Scan(&ordersToFix).Error

	if err != nil {
		return nil, fmt.Errorf("查询待修复订单失败: %w", err)
	}

	result.TotalOrders = len(ordersToFix)

	if input.DryRun {
		// 仅检查模式，不执行修复
		return result, nil
	}

	// 2. 逐个修复订单
	for _, order := range ordersToFix {
		err := s.fixSingleOrder(order.OrderID, order.ParentProfileID, order.ParentCode)
		if err != nil {
			result.FailedOrders = append(result.FailedOrders, order.OrderID)
			result.ErrorMessages = append(result.ErrorMessages,
				fmt.Sprintf("订单 %d 修复失败: %v", order.OrderID, err))
			continue
		}

		result.FixedOrders++
		result.SuccessCommissions++
	}

	return result, nil
}

// fixSingleOrder 修复单个订单的推广归因并生成分佣
func (s *AffiliateService) fixSingleOrder(orderID, parentProfileID uint, parentCode string) error {
	return s.repo.Transaction(func(tx *gorm.DB) error {
		// 1. 更新订单的推广归因
		err := tx.Model(&models.Order{}).
			Where("id = ?", orderID).
			Updates(map[string]interface{}{
				"affiliate_profile_id": parentProfileID,
				"affiliate_code":       parentCode,
				"updated_at":           time.Now(),
			}).Error

		if err != nil {
			return fmt.Errorf("更新订单归因失败: %w", err)
		}

		// 2. 触发分佣计算
		// 注意：这里需要使用事务外的 service 实例
		// 因为 HandleOrderPaid 内部会开启新事务
		return nil
	})
}

// BackfillCommissionsForOrder 为单个订单补偿分佣（事务外调用）
func (s *AffiliateService) BackfillCommissionsForOrder(orderID uint) error {
	if orderID == 0 {
		return fmt.Errorf("invalid order id")
	}

	// 1. 检查订单是否已有分佣
	var count int64
	if err := models.DB.Model(&models.AffiliateCommission{}).
		Where("order_id = ?", orderID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("订单 %d 已有分佣记录，跳过", orderID)
	}

	// 2. 检查订单状态
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != "paid" && order.Status != "completed" {
		return fmt.Errorf("订单状态不是 paid 或 completed，跳过")
	}

	// 3. 检查是否有推广归因
	if order.AffiliateProfileID == nil || *order.AffiliateProfileID == 0 {
		// 尝试自动补充归因
		var upl models.UserPromotionLevel
		if err := models.DB.Where("user_id = ?", order.UserID).First(&upl).Error; err != nil {
			return fmt.Errorf("用户无上级关系: %w", err)
		}

		if upl.ParentUserID == 0 {
			return fmt.Errorf("用户无上级")
		}

		var parentProfile models.AffiliateProfile
		if err := models.DB.Where("user_id = ?", upl.ParentUserID).First(&parentProfile).Error; err != nil {
			return fmt.Errorf("上级推广档案不存在: %w", err)
		}

		// 更新订单归因
		err := models.DB.Model(order).Updates(map[string]interface{}{
			"affiliate_profile_id": parentProfile.ID,
			"affiliate_code":       parentProfile.AffiliateCode,
			"updated_at":           time.Now(),
		}).Error

		if err != nil {
			return fmt.Errorf("更新订单归因失败: %w", err)
		}
	}

	// 4. 触发分佣计算
	return s.HandleOrderPaid(orderID)
}

// BatchBackfillCommissions 批量补偿分佣
func (s *AffiliateService) BatchBackfillCommissions(orderIDs []uint) (*BackfillCommissionsResult, error) {
	result := &BackfillCommissionsResult{
		TotalOrders:   len(orderIDs),
		FailedOrders:  []uint{},
		SkippedOrders: []uint{},
		ErrorMessages: []string{},
	}

	for _, orderID := range orderIDs {
		err := s.BackfillCommissionsForOrder(orderID)
		if err != nil {
			if err.Error() == fmt.Sprintf("订单 %d 已有分佣记录，跳过", orderID) {
				result.SkippedOrders = append(result.SkippedOrders, orderID)
			} else {
				result.FailedOrders = append(result.FailedOrders, orderID)
				result.ErrorMessages = append(result.ErrorMessages,
					fmt.Sprintf("订单 %d: %v", orderID, err))
			}
			continue
		}

		result.FixedOrders++
		result.SuccessCommissions++
	}

	return result, nil
}
