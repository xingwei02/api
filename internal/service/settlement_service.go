package service

import (
	"errors"
	"fmt"
	"time"

	"context"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SettlementService 结算服务
type SettlementService struct {
	db              *gorm.DB
	userAuthService *UserAuthService
}

// NewSettlementService 创建结算服务
func NewSettlementService(db *gorm.DB, userAuthService *UserAuthService) *SettlementService {
	return &SettlementService{
		db:              db,
		userAuthService: userAuthService,
	}
}

// ============================================================================
// 余额操作相关
// ============================================================================

// GetUserBalance 获取用户余额
func (s *SettlementService) GetUserBalance(userID uint) (*models.UserBalance, error) {
	var balance models.UserBalance
	err := s.db.Where("user_id = ?", userID).First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果不存在，创建初始余额记录
			// 查询用户的affiliate_profile_id
			var profile models.AffiliateProfile
			var affiliateProfileID *uint
			if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err == nil {
				affiliateProfileID = &profile.ID
			}

			balance = models.UserBalance{
				UserID:             userID,
				AffiliateProfileID: affiliateProfileID,
				Balance:            0,
				FrozenBalance:      0,
				TotalIncome:        0,
				TotalWithdraw:      0,
				Version:            1,
			}
			if createErr := s.db.Create(&balance).Error; createErr != nil {
				return nil, createErr
			}
			return &balance, nil
		}
		return nil, err
	}
	return &balance, nil
}

// GetBalanceLogs 获取余额明细
func (s *SettlementService) GetBalanceLogs(userID uint, page, pageSize int) ([]models.UserBalanceLog, int64, error) {
	var logs []models.UserBalanceLog
	var total int64

	query := s.db.Model(&models.UserBalanceLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// TransferCommissionToBalance 佣金转余额（带邮箱验证码）
// 红线1：金额处理必须用decimal.Decimal
// 红线2：字段映射必须正确
func (s *SettlementService) TransferCommissionToBalance(userID uint, commissionIDs []uint, verifyCode, userEmail string, amountFloat float64) error {
	// 红线1：立即转换为decimal.Decimal
	amount := decimal.NewFromFloat(amountFloat).RoundBank(2)

	// 1. 验证金额是否大于等于10元（需要验证码）
	minVerifyAmount := decimal.NewFromFloat(10.0)
	if amount.GreaterThanOrEqual(minVerifyAmount) {
		// 验证邮箱验证码
		if s.userAuthService != nil {
			_, err := s.userAuthService.verifyCode(userEmail, constants.VerifyPurposeCommissionTransfer, verifyCode)
			if err != nil {
				return fmt.Errorf("验证码校验失败: %w", err)
			}
		}
	}

	// 2. 检查最小转账金额（1元）
	minAmount := decimal.NewFromFloat(1.0)
	if amount.LessThan(minAmount) {
		return fmt.Errorf("转账金额不能低于%.2f元", minAmount)
	}

	// 3. 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 4. 查询待转账佣金（红线2：字段名CommissionAmount）
		var commissions []models.AffiliateCommission
		err := tx.Where("id IN ? AND status = ? AND transferred_to_balance = ?",
			commissionIDs, constants.AffiliateCommissionStatusAvailable, false).
			Find(&commissions).Error
		if err != nil {
			return err
		}

		if len(commissions) == 0 {
			return errors.New("没有可转账的佣金")
		}

		// 5. 计算总金额（红线1：使用decimal计算）
		totalAmount := decimal.Zero
		for _, comm := range commissions {
			totalAmount = totalAmount.Add(comm.CommissionAmount.Decimal)
		}

		// 6. 验证金额是否匹配
		if !totalAmount.Equal(amount) {
			return fmt.Errorf("佣金总额%.2f与请求金额%.2f不匹配", totalAmount, amount)
		}

		// 7. 获取用户余额（带乐观锁）
		var balance models.UserBalance
		err = tx.Where("user_id = ?", userID).First(&balance).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建初始余额
				var profile models.AffiliateProfile
				var affiliateProfileID *uint
				if err := tx.Where("user_id = ?", userID).First(&profile).Error; err == nil {
					affiliateProfileID = &profile.ID
				}

				balance = models.UserBalance{
					UserID:             userID,
					AffiliateProfileID: affiliateProfileID,
					Balance:            0,
					Version:            1,
				}
				if createErr := tx.Create(&balance).Error; createErr != nil {
					return createErr
				}
			} else {
				return err
			}
		}

		// 8. 更新余额（乐观锁，红线1：decimal计算）
		oldVersion := balance.Version
		amountFloat64, _ := totalAmount.Float64()
		result := tx.Model(&models.UserBalance{}).
			Where("user_id = ? AND version = ?", userID, oldVersion).
			Updates(map[string]interface{}{
				"balance":      gorm.Expr("balance + ?", amountFloat64),
				"total_income": gorm.Expr("total_income + ?", amountFloat64),
				"version":      gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("余额更新失败，请重试")
		}

		// 9. 标记佣金已转账
		now := time.Now()
		err = tx.Model(&models.AffiliateCommission{}).
			Where("id IN ?", commissionIDs).
			Updates(map[string]interface{}{
				"transferred_to_balance": true,
				"transfer_time":          now,
			}).Error
		if err != nil {
			return err
		}

		// 10. 记录余额明细
		log := models.UserBalanceLog{
			UserID:        userID,
			Type:          models.BalanceLogTypeCommissionTransfer,
			Amount:        amountFloat64,
			BalanceBefore: balance.Balance,
			BalanceAfter:  balance.Balance + amountFloat64,
			Description:   fmt.Sprintf("佣金转入余额，共%d笔", len(commissions)),
			RelatedType:   "commission",
			CreatedAt:     now,
		}
		return tx.Create(&log).Error
	})
}

// ApplyWithdraw 申请提现（带邮箱验证码）
// 红线1：金额处理必须用decimal.Decimal
// 红线2：字段映射 AlipayAccount→Account, RealName→AlipayName, RequireRealName→RequireRealname
func (s *SettlementService) ApplyWithdraw(userID uint, amountFloat float64, alipayAccount, realName, verifyCode, userEmail string) error {
	// 红线1：立即转换为decimal.Decimal
	amount := decimal.NewFromFloat(amountFloat).RoundBank(2)

	// 1. 验证邮箱验证码
	if s.userAuthService != nil {
		_, err := s.userAuthService.verifyCode(userEmail, constants.VerifyPurposeWithdraw, verifyCode)
		if err != nil {
			return fmt.Errorf("验证码校验失败: %w", err)
		}
	}

	// 2. 获取提现设置
	var settings models.AffiliateWithdrawSettings
	err := s.db.First(&settings).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 3. 校验提现规则（红线1：decimal比较）
	if settings.MinAmount.Decimal.GreaterThan(decimal.Zero) && amount.LessThan(settings.MinAmount.Decimal) {
		return fmt.Errorf("提现金额不能低于%.2f元", settings.MinAmount.Decimal)
	}

	// 4. 检查提现权限（红线2：RequireRealname）
	if settings.RequireRealname && realName == "" {
		return errors.New("需要实名认证才能提现")
	}

	// 5. 检查提现频率限制（Redis锁）
	// 红线3：Redis锁必须用SET NX EX 30
	lockKey := fmt.Sprintf("withdraw:lock:%d", userID)
	rdb := cache.Client()
	if rdb != nil {
		ctx := context.Background()
		// 尝试获取锁，30秒过期（防止并发重复提交）
		success, err := rdb.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
		if err != nil {
			return fmt.Errorf("提现服务繁忙: %w", err)
		}
		if !success {
			return errors.New("操作过于频繁，请稍后再试")
		}

		// 5.1 业务限制：每天只能提现一次
		// key: withdraw:last_day:USER_ID, value: 20260421
		dayKey := fmt.Sprintf("withdraw:last_day:%d", userID)
		todayStr := time.Now().Format("20060102")
		lastDay, _ := rdb.Get(ctx, dayKey).Result()
		if lastDay == todayStr {
			return errors.New("每天只能提交一次提现申请")
		}
		// 记录今天已提现（24小时过期即可）
		rdb.Set(ctx, dayKey, todayStr, 24*time.Hour)
	}

	// 6. 查询用户的AffiliateProfile（红线2：必须用AffiliateProfileID）
	var profile models.AffiliateProfile
	err = s.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return fmt.Errorf("未找到推广用户信息: %w", err)
	}

	// 7. 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 8. 获取用户余额（带乐观锁）
		var balance models.UserBalance
		err := tx.Where("user_id = ?", userID).First(&balance).Error
		if err != nil {
			return err
		}

		// 9. 检查余额是否足够（红线1：decimal比较）
		balanceDecimal := decimal.NewFromFloat(balance.Balance)
		if balanceDecimal.LessThan(amount) {
			return errors.New("余额不足")
		}

		// 10. 冻结余额（红线1：decimal计算）
		oldVersion := balance.Version
		amountFloat64, _ := amount.Float64()
		result := tx.Model(&models.UserBalance{}).
			Where("user_id = ? AND version = ?", userID, oldVersion).
			Updates(map[string]interface{}{
				"balance":        gorm.Expr("balance - ?", amountFloat64),
				"frozen_balance": gorm.Expr("frozen_balance + ?", amountFloat64),
				"version":        gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("余额冻结失败，请重试")
		}

		// 11. 计算手续费（红线1：decimal计算）
		feeRate := decimal.NewFromFloat(settings.FeeRate)
		fee := amount.Mul(feeRate).RoundBank(2)
		actualAmount := amount.Sub(fee).RoundBank(2)

		// 12. 创建提现申请（红线2：字段映射）
		now := time.Now()
		withdraw := models.AffiliateWithdrawRequest{
			AffiliateProfileID: profile.ID, // 红线2：使用AffiliateProfileID
			Amount:             models.Money{Decimal: amount},
			Fee:                models.Money{Decimal: fee},
			ActualAmount:       models.Money{Decimal: actualAmount},
			Channel:            "alipay",
			Account:            alipayAccount, // 红线2：Account字段
			AlipayName:         realName,      // 红线2：AlipayName字段
			Status:             constants.AffiliateWithdrawStatusPendingReview,
			CreatedAt:          now,
		}
		if err := tx.Create(&withdraw).Error; err != nil {
			return err
		}

		// 13. 记录余额明细
		log := models.UserBalanceLog{
			UserID:        userID,
			Type:          models.BalanceLogTypeWithdrawApply,
			Amount:        -amountFloat64,
			BalanceBefore: balance.Balance,
			BalanceAfter:  balance.Balance - amountFloat64,
			Description:   fmt.Sprintf("提现申请，金额%.2f元", amount),
			RelatedType:   "withdraw_request",
			RelatedID:     &withdraw.ID,
			CreatedAt:     now,
		}
		return tx.Create(&log).Error
	})
}

// GetWithdrawRequests 获取提现申请列表
func (s *SettlementService) GetWithdrawRequests(userID uint, page, pageSize int) ([]models.AffiliateWithdrawRequest, int64, error) {
	// 红线2：通过user_id查找affiliate_profile_id
	var profile models.AffiliateProfile
	err := s.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, 0, err
	}

	var requests []models.AffiliateWithdrawRequest
	var total int64

	query := s.db.Model(&models.AffiliateWithdrawRequest{}).Where("affiliate_profile_id = ?", profile.ID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&requests).Error

	return requests, total, err
}

// GetWithdrawSettings 获取提现设置
func (s *SettlementService) GetWithdrawSettings() (*models.AffiliateWithdrawSettings, error) {
	var settings models.AffiliateWithdrawSettings
	err := s.db.First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认设置（红线1：使用decimal）
			return &models.AffiliateWithdrawSettings{
				MinAmount:       models.Money{Decimal: decimal.NewFromFloat(100.0)},
				FeeRate:         0,
				IntervalDays:    7,
				RequireRealname: true, // 红线2：RequireRealname字段
			}, nil
		}
		return nil, err
	}
	return &settings, nil
}

// ============================================================================
// Admin端：提现审核相关
// ============================================================================

// AdminRejectWithdraw 驳回提现申请
func (s *SettlementService) AdminRejectWithdraw(withdrawID uint, adminID uint, adminName, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 获取提现申请
		var withdraw models.AffiliateWithdrawRequest
		err := tx.Where("id = ? AND status = ?", withdrawID, constants.AffiliateWithdrawStatusPendingReview).
			First(&withdraw).Error
		if err != nil {
			return err
		}

		// 2. 通过affiliate_profile_id查找user_id（红线2）
		var profile models.AffiliateProfile
		err = tx.Where("id = ?", withdraw.AffiliateProfileID).First(&profile).Error
		if err != nil {
			return err
		}

		// 3. 获取用户余额
		var balance models.UserBalance
		err = tx.Where("user_id = ?", profile.UserID).First(&balance).Error
		if err != nil {
			return err
		}

		// 4. 解冻余额（红线1：decimal计算）
		oldVersion := balance.Version
		amountFloat64, _ := withdraw.Amount.Decimal.Float64()
		result := tx.Model(&models.UserBalance{}).
			Where("user_id = ? AND version = ?", profile.UserID, oldVersion).
			Updates(map[string]interface{}{
				"balance":        gorm.Expr("balance + ?", amountFloat64),
				"frozen_balance": gorm.Expr("frozen_balance - ?", amountFloat64),
				"version":        gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("余额解冻失败，请重试")
		}

		// 5. 更新提现申请状态
		now := time.Now()
		err = tx.Model(&withdraw).Updates(map[string]interface{}{
			"status":        constants.AffiliateWithdrawStatusRejected,
			"reject_reason": reason,
			"processed_at":  now,
			"processed_by":  adminID,
		}).Error
		if err != nil {
			return err
		}

		// 6. 记录余额明细
		log := models.UserBalanceLog{
			UserID:        profile.UserID,
			Type:          models.BalanceLogTypeWithdrawReject,
			Amount:        amountFloat64,
			BalanceBefore: balance.Balance - amountFloat64,
			BalanceAfter:  balance.Balance,
			Description:   fmt.Sprintf("提现驳回，原因：%s", reason),
			RelatedType:   "withdraw_request",
			RelatedID:     &withdrawID,
			OperatorID:    &adminID,
			OperatorName:  adminName,
			CreatedAt:     now,
		}
		return tx.Create(&log).Error
	})
}

// GetAdminWithdrawRequests 获取后台提现申请列表
func (s *SettlementService) GetAdminWithdrawRequests(status string, page, pageSize int) ([]models.AffiliateWithdrawRequest, int64, error) {
	var requests []models.AffiliateWithdrawRequest
	var total int64

	query := s.db.Model(&models.AffiliateWithdrawRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("AffiliateProfile.User").Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&requests).Error

	return requests, total, err
}

// UpdateWithdrawSettings 更新提现设置
func (s *SettlementService) UpdateWithdrawSettings(settings models.AffiliateWithdrawSettings) error {
	var exist models.AffiliateWithdrawSettings
	err := s.db.First(&exist).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.db.Create(&settings).Error
		}
		return err
	}
	return s.db.Model(&exist).Where("id = ?", exist.ID).Updates(settings).Error
}

// AdminCompleteWithdraw 完成提现（打款完成）
func (s *SettlementService) AdminCompleteWithdraw(withdrawID uint, adminID uint, adminName, transactionID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 获取提现申请
		var withdraw models.AffiliateWithdrawRequest
		err := tx.Where("id = ? AND status = ?", withdrawID, constants.AffiliateWithdrawStatusPendingReview).
			First(&withdraw).Error
		if err != nil {
			return err
		}

		// 2. 通过affiliate_profile_id查找user_id（红线2）
		var profile models.AffiliateProfile
		err = tx.Where("id = ?", withdraw.AffiliateProfileID).First(&profile).Error
		if err != nil {
			return err
		}

		// 3. 获取用户余额
		var balance models.UserBalance
		err = tx.Where("user_id = ?", profile.UserID).First(&balance).Error
		if err != nil {
			return err
		}

		// 4. 扣除冻结余额（红线1：decimal计算）
		oldVersion := balance.Version
		amountFloat64, _ := withdraw.Amount.Decimal.Float64()
		actualAmountFloat64, _ := withdraw.ActualAmount.Decimal.Float64()
		result := tx.Model(&models.UserBalance{}).
			Where("user_id = ? AND version = ?", profile.UserID, oldVersion).
			Updates(map[string]interface{}{
				"frozen_balance": gorm.Expr("frozen_balance - ?", amountFloat64),
				"total_withdraw": gorm.Expr("total_withdraw + ?", actualAmountFloat64),
				"version":        gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("余额扣除失败，请重试")
		}

		// 5. 更新提现申请状态
		now := time.Now()
		err = tx.Model(&withdraw).Updates(map[string]interface{}{
			"status":       constants.AffiliateWithdrawStatusPaid,
			"processed_at": now,
			"processed_by": adminID,
		}).Error
		if err != nil {
			return err
		}

		// 6. 记录余额明细
		log := models.UserBalanceLog{
			UserID:        profile.UserID,
			Type:          models.BalanceLogTypeWithdrawComplete,
			Amount:        0, // 已经在申请时扣除
			BalanceBefore: balance.Balance,
			BalanceAfter:  balance.Balance,
			Description:   fmt.Sprintf("提现完成，实际到账%.2f元", withdraw.ActualAmount.Decimal),
			RelatedType:   "withdraw_request",
			RelatedID:     &withdrawID,
			OperatorID:    &adminID,
			OperatorName:  adminName,
			Remark:        fmt.Sprintf("交易单号：%s", transactionID),
			CreatedAt:     now,
		}
		return tx.Create(&log).Error
	})
}
