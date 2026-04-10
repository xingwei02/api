package service

import (
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"gorm.io/gorm"
)

// OrderServiceWithPromotion OrderService 的推广方案扩展
// 这个文件展示如何在现有的 OrderService 中集成 PromotionService

// 在 OrderService 结构体中添加：
// promotionSvc *PromotionService

// InitOrderServiceWithPromotion 初始化 OrderService 时添加推广服务
func (s *OrderService) InitOrderServiceWithPromotion(promotionSvc *PromotionService) {
	// 在 OrderService 中添加字段：promotionSvc *PromotionService
	// s.promotionSvc = promotionSvc
}

// handleOrderPaidWithPromotion 订单支付完成时处理推广逻辑
// 这个方法应该在 updateOrderToPaidInTx 或 ConfirmOrderStatus 中调用
func (s *OrderService) handleOrderPaidWithPromotion(tx *gorm.DB, order *models.Order) error {
	if s.promotionSvc == nil || order == nil || order.UserID == 0 {
		return nil
	}

	// 1. 记录考核周期数据
	if err := s.promotionSvc.RecordCycleData(order.UserID, order.TotalAmount, 1); err != nil {
		// 记录错误但不中断订单流程
		logger.Warnw("promotion_record_cycle_data_failed",
			"order_id", order.ID,
			"user_id", order.UserID,
			"error", err,
		)
	}

	// 2. 检查是否满足升级条件
	shouldUpgrade, err := s.promotionSvc.CheckUpgrade(order.UserID)
	if err != nil {
		logger.Warnw("promotion_check_upgrade_failed",
			"order_id", order.ID,
			"user_id", order.UserID,
			"error", err,
		)
	} else if shouldUpgrade {
		// 3. 执行升级
		if err := s.promotionSvc.UpdateUserLevel(order.UserID); err != nil {
			logger.Warnw("promotion_update_level_failed",
				"order_id", order.ID,
				"user_id", order.UserID,
				"error", err,
			)
		}
	}

	return nil
}

// 集成点说明：
// 在 order_service_child.go 的 updateOrderToPaidInTx 函数中，
// 在调用 s.affiliateSvc.HandleOrderPaid(order.ID) 之后添加：
//
// if s.promotionSvc != nil {
//     if err := s.handleOrderPaidWithPromotion(tx, order); err != nil {
//         logger.Warnw("promotion_handle_order_paid_failed",
//             "order_id", order.ID,
//             "error", err,
//         )
//     }
// }
//
// 同时在 ConfirmOrderStatus 函数中的订单支付完成分支也要添加相同的调用。

// GetUserPromotionInfo 获取用户的推广信息（用于 API 响应）
func (s *OrderService) GetUserPromotionInfo(userID uint) (map[string]interface{}, error) {
	if s.promotionSvc == nil {
		return nil, nil
	}

	progress, err := s.promotionSvc.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	return progress, nil
}

// InitializeUserPromotion 初始化用户推广（当用户被邀请时调用）
func (s *OrderService) InitializeUserPromotion(userID, parentUserID uint) error {
	if s.promotionSvc == nil {
		return nil
	}

	return s.promotionSvc.InitializeUserLevel(userID, parentUserID)
}
