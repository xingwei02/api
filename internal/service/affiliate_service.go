package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	affiliateCodeLength        = 8
	affiliateSplitTypePrefix   = "sp"
	affiliateAttributionWindow = 30 * 24 * time.Hour
	affiliateClickDedupeWindow = 10 * time.Minute
)

// AffiliateService 推广返利业务服务
type AffiliateService struct {
	repo           repository.AffiliateRepository
	userRepo       repository.UserRepository
	orderRepo      repository.OrderRepository
	productRepo    repository.ProductRepository
	settingService *SettingService
}

// NewAffiliateService 创建推广返利服务
func NewAffiliateService(
	repo repository.AffiliateRepository,
	userRepo repository.UserRepository,
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	settingService *SettingService,
) *AffiliateService {
	return &AffiliateService{
		repo:           repo,
		userRepo:       userRepo,
		orderRepo:      orderRepo,
		productRepo:    productRepo,
		settingService: settingService,
	}
}

// AffiliateTrackClickInput 推广点击记录输入
type AffiliateTrackClickInput struct {
	AffiliateCode string
	VisitorKey    string
	LandingPath   string
	Referrer      string
	ClientIP      string
	UserAgent     string
}

// AffiliateDashboard 推广用户中心数据
type AffiliateDashboard struct {
	Opened              bool         `json:"opened"`
	AffiliateCode       string       `json:"affiliate_code"`
	PromotionPath       string       `json:"promotion_path"`
	ClickCount          int64        `json:"click_count"`
	ValidOrderCount     int64        `json:"valid_order_count"`
	ConversionRate      float64      `json:"conversion_rate"`
	PendingCommission   models.Money `json:"pending_commission"`
	AvailableCommission models.Money `json:"available_commission"`
	WithdrawnCommission models.Money `json:"withdrawn_commission"`
}

// AffiliateStats 推广统计数据
type AffiliateStats struct {
	ClickCount          int64
	ValidOrderCount     int64
	ConversionRate      float64
	PendingCommission   models.Money
	AvailableCommission models.Money
	WithdrawnCommission models.Money
}

// AffiliateWithdrawApplyInput 提现申请输入
type AffiliateWithdrawApplyInput struct {
	Amount  decimal.Decimal
	Channel string
	Account string
}

// AffiliateAdminUserItem 后台推广用户列表项
type AffiliateAdminUserItem struct {
	Profile           models.AffiliateProfile `json:"profile"`
	Stats             AffiliateStats          `json:"stats"`
	TopDiscountRate   float64                 `json:"top_discount_rate"`
	HasParentPromoter bool                    `json:"has_parent_promoter"`
}

// AffiliateAdminCommissionListFilter 后台佣金列表过滤
type AffiliateAdminCommissionListFilter struct {
	Page               int
	PageSize           int
	AffiliateProfileID uint
	OrderNo            string
	Status             string
	Keyword            string
}

// AffiliateAdminWithdrawListFilter 后台提现列表过滤
type AffiliateAdminWithdrawListFilter struct {
	Page               int
	PageSize           int
	AffiliateProfileID uint
	Status             string
	Keyword            string
}

// UpdateAffiliateProfileStatus 管理端更新返利用户状态
func (s *AffiliateService) UpdateAffiliateProfileStatus(profileID uint, rawStatus string) (*models.AffiliateProfile, error) {
	if profileID == 0 || s.repo == nil {
		return nil, ErrNotFound
	}
	nextStatus := strings.TrimSpace(rawStatus)
	if nextStatus != constants.AffiliateProfileStatusActive && nextStatus != constants.AffiliateProfileStatusDisabled {
		return nil, ErrAffiliateProfileStatusInvalid
	}

	profile, err := s.repo.GetProfileByID(profileID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrNotFound
	}
	if strings.TrimSpace(profile.Status) == nextStatus {
		return profile, nil
	}
	if err := s.repo.UpdateProfileStatus(profileID, nextStatus, time.Now()); err != nil {
		return nil, err
	}
	return s.repo.GetProfileByID(profileID)
}

// BatchUpdateAffiliateProfileStatus 管理端批量更新返利用户状态
func (s *AffiliateService) BatchUpdateAffiliateProfileStatus(profileIDs []uint, rawStatus string) (int64, error) {
	if s.repo == nil {
		return 0, ErrNotFound
	}
	nextStatus := strings.TrimSpace(rawStatus)
	if nextStatus != constants.AffiliateProfileStatusActive && nextStatus != constants.AffiliateProfileStatusDisabled {
		return 0, ErrAffiliateProfileStatusInvalid
	}
	normalizedIDs := normalizeAffiliateProfileIDs(profileIDs)
	if len(normalizedIDs) == 0 {
		return 0, nil
	}
	return s.repo.BatchUpdateProfileStatus(normalizedIDs, nextStatus, time.Now())
}

// OpenAffiliate 为用户开通推广返利
func (s *AffiliateService) OpenAffiliate(userID uint) (*models.AffiliateProfile, error) {
	if userID == 0 {
		return nil, ErrUserDisabled
	}
	if s.repo == nil || s.userRepo == nil {
		return nil, ErrNotFound
	}
	setting := AffiliateDefaultSetting()
	var err error
	if s.settingService != nil {
		setting, err = s.settingService.GetAffiliateSetting()
		if err != nil {
			return nil, err
		}
	}
	if !setting.Enabled {
		return nil, ErrAffiliateDisabled
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	if strings.TrimSpace(user.Status) == constants.UserStatusDisabled {
		return nil, ErrUserDisabled
	}

	existing, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	const maxRetry = 8
	for i := 0; i < maxRetry; i++ {
		code, genErr := generateAffiliateCode()
		if genErr != nil {
			return nil, genErr
		}
		profile := &models.AffiliateProfile{
			UserID:        userID,
			AffiliateCode: code,
			Status:        constants.AffiliateProfileStatusActive,
		}
		if err := s.repo.CreateProfile(profile); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return nil, err
		}
		created, err := s.repo.GetProfileByID(profile.ID)
		if err != nil {
			return nil, err
		}
		if created != nil {
			return created, nil
		}
		return profile, nil
	}
	return nil, ErrAffiliateCodeInvalid
}

// OpenTokenMerchant 开通 Token 商身份，并尽量挂接邀请关系与默认方案。
// 注意：Token 商开通是独立入口，不受推广返利总开关（setting.Enabled）限制。
func (s *AffiliateService) OpenTokenMerchant(userID uint, inviterCode string) (*models.AffiliateProfile, error) {
	if userID == 0 || s.repo == nil || s.userRepo == nil {
		return nil, ErrNotFound
	}
	// 直接获取或创建 AffiliateProfile，不经过总开关检查
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		// 用户不存在时先校验
		u, err := s.userRepo.GetByID(userID)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, ErrNotFound
		}
		if strings.TrimSpace(u.Status) == constants.UserStatusDisabled {
			return nil, ErrUserDisabled
		}
		const maxRetry = 8
		for i := 0; i < maxRetry; i++ {
			code, genErr := generateAffiliateCode()
			if genErr != nil {
				return nil, genErr
			}
			p := &models.AffiliateProfile{
				UserID:        userID,
				AffiliateCode: code,
				Status:        constants.AffiliateProfileStatusActive,
			}
			if err := s.repo.CreateProfile(p); err != nil {
				if isUniqueViolation(err) {
					continue
				}
				return nil, err
			}
			created, err := s.repo.GetProfileByID(p.ID)
			if err != nil {
				return nil, err
			}
			if created != nil {
				profile = created
			} else {
				profile = p
			}
			break
		}
		if profile == nil {
			return nil, ErrAffiliateCodeInvalid
		}
	}
	if s.userRepo == nil {
		return profile, nil
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	now := time.Now()
	if !user.IsTokenMerchant {
		user.IsTokenMerchant = true
		user.TokenMerchantAt = &now
		if err := s.userRepo.Update(user); err != nil {
			return nil, err
		}
		_ = cache.SetUserAuthState(context.Background(), cache.BuildUserAuthState(user))
	}
	if err := s.ensureTokenMerchantScheme(userID); err != nil {
		return nil, err
	}
	if err := s.bindTokenMerchantInviter(userID, inviterCode); err != nil {
		return nil, err
	}
	return profile, nil
}

// GetPublicAffiliateProfileByCode 按联盟ID获取公开可用的推广档案
func (s *AffiliateService) GetPublicAffiliateProfileByCode(code string) (*models.AffiliateProfile, error) {
	if s == nil || s.repo == nil {
		return nil, ErrNotFound
	}
	profile, err := s.repo.GetProfileByCode(code)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}
	if strings.TrimSpace(profile.Status) != constants.AffiliateProfileStatusActive {
		return nil, nil
	}
	return profile, nil
}

func (s *AffiliateService) ensureTokenMerchantScheme(userID uint) error {
	if userID == 0 || models.DB == nil {
		return nil
	}
	var scheme models.AffiliateLevelScheme
	err := models.DB.Where("user_id = ?", userID).First(&scheme).Error
	if err == nil {
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	scheme = models.AffiliateLevelScheme{
		UserID:    userID,
		MyRate:    0,
		EntryRate: 0,
		Version:   1,
	}
	return models.DB.Create(&scheme).Error
}

func (s *AffiliateService) bindTokenMerchantInviter(userID uint, inviterCode string) error {
	code := normalizeAffiliateCode(inviterCode)
	if userID == 0 || code == "" || s.repo == nil || models.DB == nil {
		return nil
	}
	inviterProfile, err := s.repo.GetProfileByCode(code)
	if err != nil {
		return err
	}
	if inviterProfile == nil || inviterProfile.UserID == 0 || inviterProfile.UserID == userID {
		return nil
	}

	var existing models.UserPromotionLevel
	err = models.DB.Where("user_id = ?", userID).First(&existing).Error
	if err == nil {
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	entryRate := 0.0
	entryItemID := uint(0)
	var inviterScheme models.AffiliateLevelScheme
	if err := models.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order asc, id asc")
	}).Where("user_id = ?", inviterProfile.UserID).First(&inviterScheme).Error; err == nil {
		entryRate = inviterScheme.EntryRate
		for _, item := range inviterScheme.Items {
			if item.IsEntry {
				entryItemID = item.ID
				if item.Rate > 0 {
					entryRate = item.Rate
				}
				break
			}
		}
	}

	now := time.Now()
	level := models.UserPromotionLevel{
		UserID:       userID,
		ParentUserID: inviterProfile.UserID,
		LevelItemID:  entryItemID,
		MaxRate:      entryRate,
		CustomRate:   -1,
		CurrentLevel: 1,
		CurrentRate:  entryRate,
		CycleStart:   now,
		CycleEnd:     now.AddDate(0, 1, 0),
	}
	return models.DB.Create(&level).Error
}

// ResolveOrderAffiliateSnapshot 解析下单归因快照（最近30天最后一次有效点击优先）
func (s *AffiliateService) ResolveOrderAffiliateSnapshot(userID uint, rawCode, rawVisitorKey string) (*uint, string, error) {
	code := normalizeAffiliateCode(rawCode)
	visitorKey := strings.TrimSpace(rawVisitorKey)
	if s.repo == nil {
		return nil, "", nil
	}

	setting, err := s.settingService.GetAffiliateSetting()
	if err != nil {
		return nil, "", err
	}

	// visitor_key 优先（最近30天最后一次有效点击）
	if visitorKey != "" {
		profile, err := s.repo.GetLatestActiveProfileByVisitorKey(visitorKey, time.Now().Add(-affiliateAttributionWindow))
		if err != nil {
			return nil, "", err
		}
		if profile != nil {
			if userID != 0 && profile.UserID == userID {
				return nil, "", nil
			}
			if !setting.Enabled && !s.isTokenMerchantUser(profile.UserID) {
				return nil, "", nil
			}
			profileID := profile.ID
			return &profileID, profile.AffiliateCode, nil
		}
	}

	if code == "" {
		return s.resolveAffiliateSnapshotFromPromotionRelation(setting, userID)
	}

	profile, err := s.repo.GetProfileByCode(code)
	if err != nil {
		return nil, "", err
	}
	if profile == nil || strings.TrimSpace(profile.Status) != constants.AffiliateProfileStatusActive {
		return s.resolveAffiliateSnapshotFromPromotionRelation(setting, userID)
	}
	if userID != 0 && profile.UserID == userID {
		return s.resolveAffiliateSnapshotFromPromotionRelation(setting, userID)
	}
	if !setting.Enabled && !s.isTokenMerchantUser(profile.UserID) {
		return s.resolveAffiliateSnapshotFromPromotionRelation(setting, userID)
	}

	profileID := profile.ID
	return &profileID, profile.AffiliateCode, nil
}

func (s *AffiliateService) resolveAffiliateSnapshotFromPromotionRelation(setting AffiliateSetting, userID uint) (*uint, string, error) {
	// 无点击/邀请码命中时，已建立上下级关系的注册用户仍应稳定归属到直属上级，
	// 否则订单虽然存在，但不会进入推广统计与返佣事实链。
	if userID == 0 || models.DB == nil {
		return nil, "", nil
	}
	profile, err := s.resolveAffiliateProfileByPromotionRelation(models.DB, userID)
	if err != nil {
		return nil, "", err
	}
	if profile == nil {
		return nil, "", nil
	}
	if !setting.Enabled && !s.isTokenMerchantUser(profile.UserID) {
		return nil, "", nil
	}
	profileID := profile.ID
	return &profileID, profile.AffiliateCode, nil
}

// TrackClick 记录推广点击
func (s *AffiliateService) TrackClick(input AffiliateTrackClickInput) error {
	if s.repo == nil {
		return nil
	}
	code := normalizeAffiliateCode(input.AffiliateCode)
	if code == "" {
		return nil
	}
	setting, err := s.settingService.GetAffiliateSetting()
	if err != nil {
		return err
	}
	if !setting.Enabled {
		return nil
	}
	profile, err := s.repo.GetProfileByCode(code)
	if err != nil {
		return err
	}
	if profile == nil || strings.TrimSpace(profile.Status) != constants.AffiliateProfileStatusActive {
		return nil
	}
	visitorKey := strings.TrimSpace(input.VisitorKey)
	landingPath := strings.TrimSpace(input.LandingPath)
	if visitorKey != "" {
		duplicated, err := s.repo.HasRecentClick(profile.ID, visitorKey, landingPath, time.Now().Add(-affiliateClickDedupeWindow))
		if err != nil {
			return err
		}
		if duplicated {
			return nil
		}
	}

	click := &models.AffiliateClick{
		AffiliateProfileID: profile.ID,
		VisitorKey:         visitorKey,
		LandingPath:        landingPath,
		Referrer:           strings.TrimSpace(input.Referrer),
		ClientIP:           strings.TrimSpace(input.ClientIP),
		UserAgent:          strings.TrimSpace(input.UserAgent),
		CreatedAt:          time.Now(),
	}
	return s.repo.CreateClick(click)
}

// HandleOrderPaid 处理订单支付成功后的佣金生成（多级分佣）
//
// 业务规则（最终版）：
//  1. 分佣基数 = 原价（OriginalAmount），不是实付金额
//  2. 直接推广者佣金 = 原价 × 直接档位 - 优惠金额
//  3. 中间上级佣金 = 原价 × (本级生效档位 - 下一级生效档位)
//  4. 顶层商户佣金 = 原价 × (顶层总代上限 - 最后一层生效档位)
//  5. 优惠仅由直接推广者承担
//  6. 顶层总代上限动态读取 affiliate_level_schemes.my_rate，不写死 50%
//  7. 待结算余额只在“手动转入余额”前累计，7 天后仅修改佣金状态，不自动转钱
//  8. 同时写入 commission_layers 层级表 + user_balances 待结算余额
func (s *AffiliateService) HandleOrderPaid(orderID uint) error {
	if orderID == 0 || s.repo == nil || s.orderRepo == nil {
		return nil
	}

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return nil
	}

	// 找到直接推广者的 profile
	profile, err := s.resolveAffiliateProfileForOrder(order)
	if err != nil {
		return err
	}
	if profile == nil {
		return nil
	}
	if strings.TrimSpace(profile.Status) != constants.AffiliateProfileStatusActive {
		return nil
	}

	setting, err := s.settingService.GetAffiliateSetting()
	if err != nil {
		return err
	}
	if !setting.Enabled && !s.isTokenMerchantUser(profile.UserID) {
		return nil
	}

	// ========== 核心修复：分佣基数用原价，不是实付 ==========
	originalAmount := order.OriginalAmount.Decimal.Round(2)
	if originalAmount.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	// 优惠金额 = 原价 - 实付
	discountAmount := originalAmount.Sub(order.TotalAmount.Decimal.Round(2))
	if discountAmount.LessThan(decimal.Zero) {
		discountAmount = decimal.Zero
	}

	paidAt := time.Now()
	if order.PaidAt != nil {
		paidAt = *order.PaidAt
	}
	confirmDays := setting.ConfirmDays

	// ========== 事务：写佣金记录 + 层级表 + 待结算余额 ==========
	return s.repo.Transaction(func(tx *gorm.DB) error {
		repoTx := s.repo.WithTx(tx)
		if err := s.syncOrderAffiliateSnapshotTx(tx, order, profile); err != nil {
			return err
		}

		chain, topRate, err := s.buildCommissionChain(tx, profile.UserID)
		if err != nil {
			return err
		}
		if len(chain) == 0 {
			return nil
		}
		directNode := chain[0]
		if directNode.ProfileID != profile.ID {
			// 正常情况下应一致；这里以订单归因到的 profile 为准，避免脏数据导致层级错绑。
			directNode.ProfileID = profile.ID
		}
		hasAffiliateAttribution := order.AffiliateProfileID != nil && *order.AffiliateProfileID != 0
		if !hasAffiliateAttribution && directNode.Rate.LessThanOrEqual(decimal.Zero) {
			return nil
		}
		if topRate.LessThan(directNode.Rate) {
			topRate = directNode.Rate
		}

		layerNum := 1

		createCommissionWithLayer := func(
			affiliateProfileID uint,
			userID uint,
			rate decimal.Decimal,
			amount decimal.Decimal,
			role string,
		) error {
			persistZeroDirect := hasAffiliateAttribution && role == models.CommissionRoleDirect
			if amount.LessThan(decimal.Zero) {
				return nil
			}
			if amount.Equal(decimal.Zero) && !persistZeroDirect {
				return nil
			}

			existing, err2 := repoTx.GetCommissionByOrderAndProfile(order.ID, affiliateProfileID, constants.AffiliateCommissionTypeOrder)
			if err2 != nil {
				return err2
			}
			if existing != nil {
				return nil
			}

			status := constants.AffiliateCommissionStatusPendingConfirm
			var confirmAt *time.Time
			var availableAt *time.Time
			if confirmDays <= 0 {
				status = constants.AffiliateCommissionStatusAvailable
				availableAt = &paidAt
			} else {
				t := paidAt.Add(time.Duration(confirmDays) * 24 * time.Hour)
				confirmAt = &t
			}

			c := &models.AffiliateCommission{
				AffiliateProfileID: affiliateProfileID,
				OrderID:            order.ID,
				CommissionType:     constants.AffiliateCommissionTypeOrder,
				BaseAmount:         models.NewMoneyFromDecimal(originalAmount),
				RatePercent:        models.NewMoneyFromDecimal(rate),
				CommissionAmount:   models.NewMoneyFromDecimal(amount),
				Status:             status,
				ConfirmAt:          confirmAt,
				AvailableAt:        availableAt,
			}
			if err := repoTx.CreateCommission(c); err != nil {
				return err
			}

			layer := &models.OrderCommissionLayer{
				OrderID:            order.ID,
				LayerNum:           layerNum,
				UserID:             userID,
				AffiliateProfileID: affiliateProfileID,
				Role:               role,
				Rate:               models.NewMoneyFromDecimal(rate),
				Amount:             models.NewMoneyFromDecimal(amount),
				CommissionID:       &c.ID,
			}
			if err := tx.Create(layer).Error; err != nil {
				return err
			}

			if amount.LessThanOrEqual(decimal.Zero) {
				layerNum++
				return nil
			}

			if !hasUserBalanceLedgerTables(tx) {
				layerNum++
				return nil
			}

			var ub models.UserBalance
			if err := tx.Where("user_id = ?", userID).First(&ub).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					ub = models.UserBalance{
						UserID:             userID,
						AffiliateProfileID: &affiliateProfileID,
						PendingBalance:     amount.InexactFloat64(),
						Version:            1,
					}
					if err := tx.Create(&ub).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				result := tx.Model(&models.UserBalance{}).
					Where("id = ? AND version = ?", ub.ID, ub.Version).
					Updates(map[string]interface{}{
						"pending_balance": gorm.Expr("pending_balance + ?", amount.InexactFloat64()),
						"version":         gorm.Expr("version + 1"),
					})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("余额乐观锁冲突，用户%d", userID)
				}
			}

			balanceBefore := ub.PendingBalance - amount.InexactFloat64()
			balanceAfter := ub.PendingBalance
			if ub.ID == 0 {
				balanceBefore = 0
				balanceAfter = amount.InexactFloat64()
			} else {
				balanceBefore = ub.PendingBalance
				balanceAfter = ub.PendingBalance + amount.InexactFloat64()
			}

			balanceLog := &models.UserBalanceLog{
				UserID:        userID,
				Type:          models.BalanceLogTypeCommissionPending,
				Amount:        amount.InexactFloat64(),
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceAfter,
				Description:   fmt.Sprintf("订单%s佣金待结算", order.OrderNo),
				RelatedType:   "order_commission",
				RelatedID:     &c.ID,
			}
			if err := tx.Create(balanceLog).Error; err != nil {
				return err
			}

			layerNum++
			return nil
		}

		// ========== 1. 直接推广者佣金 = 原价×档位比例 - 优惠金额 ==========
		directGrossAmount := originalAmount.Mul(directNode.Rate).Div(decimal.NewFromInt(100)).Round(2)
		directAmount := directGrossAmount.Sub(discountAmount)
		if directAmount.LessThan(decimal.Zero) {
			directAmount = decimal.Zero
		}
		if err := createCommissionWithLayer(directNode.ProfileID, directNode.UserID, directNode.Rate, directAmount, models.CommissionRoleDirect); err != nil {
			return err
		}

		// ========== 2. 中间上级逐级差价佣金（按 user_promotion_levels.current_rate） ==========
		for i := 1; i < len(chain); i++ {
			parentNode := chain[i]
			childNode := chain[i-1]
			diff := parentNode.Rate.Sub(childNode.Rate).Round(2)
			if diff.GreaterThan(decimal.Zero) {
				diffAmount := originalAmount.Mul(diff).Div(decimal.NewFromInt(100)).Round(2)
				if err := createCommissionWithLayer(parentNode.ProfileID, parentNode.UserID, diff, diffAmount, models.CommissionRoleIndirect); err != nil {
					return err
				}
			}
		}

		// ========== 3. 顶层商户剩余差价佣金（总代上限 - 链顶最后一层） ==========
		lastRate := chain[len(chain)-1].Rate
		topDiff := topRate.Sub(lastRate).Round(2)
		if topDiff.GreaterThan(decimal.Zero) {
			topProfile, err := s.getActiveAffiliateProfileByUserID(tx, chain[len(chain)-1].UserID)
			if err != nil {
				return err
			}
			if topProfile != nil {
				topAmount := originalAmount.Mul(topDiff).Div(decimal.NewFromInt(100)).Round(2)
				if err := createCommissionWithLayer(topProfile.ID, topProfile.UserID, topDiff, topAmount, models.CommissionRoleIndirect); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// ConfirmDueCommissions 将到期佣金转可提现
func (s *AffiliateService) ConfirmDueCommissions(now time.Time) error {
	if s.repo == nil {
		return nil
	}
	_, err := s.repo.MarkPendingCommissionsAvailable(now, now)
	return err
}

// HandleOrderCanceled 处理订单取消/退款后的佣金逆向
func (s *AffiliateService) HandleOrderCanceled(orderID uint, reason string) error {
	if orderID == 0 || s.repo == nil {
		return nil
	}
	rows, err := s.repo.ListCommissionsByOrder(orderID, []string{
		constants.AffiliateCommissionStatusPendingConfirm,
		constants.AffiliateCommissionStatusAvailable,
	})
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	now := time.Now()
	reasonText := strings.TrimSpace(reason)
	if reasonText == "" {
		reasonText = "order_canceled"
	}
	for i := range rows {
		item := rows[i]
		if item.TransferredToBalance {
			// 已手动转入余额的佣金，退款/取消不再处理。
			continue
		}
		if item.WithdrawRequestID != nil {
			// 已进入提现流程，按业务规则不影响用户提现。
			continue
		}
		item.Status = constants.AffiliateCommissionStatusRejected
		item.InvalidReason = reasonText
		item.UpdatedAt = now
		if err := s.repo.UpdateCommission(&item); err != nil {
			return err
		}
	}
	return nil
}

// HandleOrderRefundedTx 在事务内处理订单退款后的佣金回滚
func (s *AffiliateService) HandleOrderRefundedTx(
	tx *gorm.DB,
	order *models.Order,
	refundDelta decimal.Decimal,
	refundedBefore decimal.Decimal,
	reason string,
) error {
	if tx == nil || order == nil || order.ID == 0 || s.repo == nil {
		return nil
	}
	delta := refundDelta.Round(2)
	if delta.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	totalAmount := order.TotalAmount.Decimal.Round(2)
	if totalAmount.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	before := refundedBefore.Round(2)
	if before.LessThan(decimal.Zero) {
		before = decimal.Zero
	}
	if before.GreaterThan(totalAmount) {
		before = totalAmount
	}
	remaining := totalAmount.Sub(before).Round(2)
	if remaining.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	if delta.GreaterThan(remaining) {
		delta = remaining
	}

	repoTx := s.repo.WithTx(tx)
	rows, err := repoTx.ListCommissionsByOrderForUpdate(order.ID, []string{
		constants.AffiliateCommissionStatusPendingConfirm,
		constants.AffiliateCommissionStatusAvailable,
	})
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	now := time.Now()
	reasonText := strings.TrimSpace(reason)
	if reasonText == "" {
		reasonText = "order_refunded"
	}
	for i := range rows {
		item := rows[i]
		if item.TransferredToBalance {
			// 已手动转入余额的佣金，退款/取消不再处理。
			continue
		}
		if item.WithdrawRequestID != nil {
			// 已进入提现流程，按业务规则不影响用户提现。
			continue
		}

		currentCommission := item.CommissionAmount.Decimal.Round(2)
		if currentCommission.LessThanOrEqual(decimal.Zero) {
			item.Status = constants.AffiliateCommissionStatusRejected
			item.InvalidReason = reasonText
			item.ConfirmAt = nil
			item.AvailableAt = nil
			item.UpdatedAt = now
			if err := repoTx.UpdateCommission(&item); err != nil {
				return err
			}
			continue
		}

		// 按“本次退款金额 / 当前剩余未退款金额”比例扣减当前佣金，避免多次退款时重复放大扣减。
		deduct := currentCommission.Mul(delta).Div(remaining).Round(2)
		nextCommission := currentCommission.Sub(deduct).Round(2)
		if nextCommission.LessThan(decimal.Zero) {
			nextCommission = decimal.Zero
		}
		currentBase := item.BaseAmount.Decimal.Round(2)
		nextBase := currentBase
		if currentBase.GreaterThan(decimal.Zero) {
			baseDeduct := currentBase.Mul(delta).Div(remaining).Round(2)
			nextBase = currentBase.Sub(baseDeduct).Round(2)
			if nextBase.LessThan(decimal.Zero) {
				nextBase = decimal.Zero
			}
		}

		item.CommissionAmount = models.NewMoneyFromDecimal(nextCommission)
		item.BaseAmount = models.NewMoneyFromDecimal(nextBase)
		item.UpdatedAt = now

		var profile models.AffiliateProfile
		if hasUserBalanceLedgerTables(tx) && tx.Select("id", "user_id").First(&profile, item.AffiliateProfileID).Error == nil && profile.UserID != 0 {
			pendingDeduct := deduct.InexactFloat64()
			var balance models.UserBalance
			if err := tx.Where("user_id = ?", profile.UserID).First(&balance).Error; err == nil {
				newPending := decimal.NewFromFloat(balance.PendingBalance).Sub(deduct).Round(2)
				if newPending.LessThan(decimal.Zero) {
					newPending = decimal.Zero
				}
				result := tx.Model(&models.UserBalance{}).
					Where("id = ? AND version = ?", balance.ID, balance.Version).
					Updates(map[string]interface{}{
						"pending_balance": newPending.InexactFloat64(),
						"version":         gorm.Expr("version + 1"),
					})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("余额乐观锁冲突，用户%d", profile.UserID)
				}

				balanceLog := &models.UserBalanceLog{
					UserID:        profile.UserID,
					Type:          models.BalanceLogTypeRefund,
					Amount:        -pendingDeduct,
					BalanceBefore: balance.PendingBalance,
					BalanceAfter:  newPending.InexactFloat64(),
					Description:   fmt.Sprintf("订单%s退款回滚待结算佣金", order.OrderNo),
					RelatedType:   "order_commission",
					RelatedID:     &item.ID,
				}
				if err := tx.Create(balanceLog).Error; err != nil {
					return err
				}
			}
		}

		if nextCommission.LessThanOrEqual(decimal.Zero) {
			item.Status = constants.AffiliateCommissionStatusRejected
			item.InvalidReason = reasonText
			item.ConfirmAt = nil
			item.AvailableAt = nil
		}
		if err := repoTx.UpdateCommission(&item); err != nil {
			return err
		}
	}
	return nil
}

func hasUserBalanceLedgerTables(tx *gorm.DB) bool {
	if tx == nil || tx.Migrator() == nil {
		return false
	}
	return tx.Migrator().HasTable(&models.UserBalance{}) && tx.Migrator().HasTable(&models.UserBalanceLog{})
}

type affiliateCommissionChainNode struct {
	UserID    uint
	ProfileID uint
	Rate      decimal.Decimal
}

func (s *AffiliateService) buildCommissionChain(tx *gorm.DB, directUserID uint) ([]affiliateCommissionChainNode, decimal.Decimal, error) {
	if tx == nil || directUserID == 0 {
		return nil, decimal.Zero, nil
	}

	chain := make([]affiliateCommissionChainNode, 0, 10)
	visited := make(map[uint]struct{}, 10)
	currentUserID := directUserID

	for i := 0; i < 10 && currentUserID != 0; i++ {
		if _, ok := visited[currentUserID]; ok {
			break
		}
		visited[currentUserID] = struct{}{}

		profile, err := s.getActiveAffiliateProfileByUserID(tx, currentUserID)
		if err != nil {
			return nil, decimal.Zero, err
		}
		if profile == nil {
			break
		}

		rate, level, err := s.getEffectivePromotionRate(tx, currentUserID)
		if err != nil {
			return nil, decimal.Zero, err
		}
		if rate.LessThan(decimal.Zero) {
			rate = decimal.Zero
		}

		chain = append(chain, affiliateCommissionChainNode{
			UserID:    currentUserID,
			ProfileID: profile.ID,
			Rate:      rate,
		})

		if level == nil || level.ParentUserID == 0 {
			break
		}
		currentUserID = level.ParentUserID
	}

	if len(chain) == 0 {
		return chain, decimal.Zero, nil
	}

	topRate, err := s.getTopMerchantRate(tx, chain[len(chain)-1].UserID)
	if err != nil {
		return nil, decimal.Zero, err
	}
	return chain, topRate, nil
}

func (s *AffiliateService) getEffectivePromotionRate(tx *gorm.DB, userID uint) (decimal.Decimal, *models.UserPromotionLevel, error) {
	if tx == nil || userID == 0 {
		return decimal.Zero, nil, nil
	}
	var level models.UserPromotionLevel
	if err := tx.Where("user_id = ?", userID).First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var scheme models.AffiliateLevelScheme
			if err := tx.Where("user_id = ?", userID).First(&scheme).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return decimal.Zero, nil, nil
				}
				return decimal.Zero, nil, err
			}
			return decimal.NewFromFloat(scheme.MyRate).Round(2), nil, nil
		}
		return decimal.Zero, nil, err
	}
	return decimal.NewFromFloat(level.CurrentRate).Round(2), &level, nil
}

func (s *AffiliateService) getTopMerchantRate(tx *gorm.DB, userID uint) (decimal.Decimal, error) {
	if tx == nil || userID == 0 {
		return decimal.Zero, nil
	}

	if rate, _, err := s.getEffectivePromotionRate(tx, userID); err != nil {
		return decimal.Zero, err
	} else if rate.GreaterThan(decimal.Zero) {
		var scheme models.AffiliateLevelScheme
		if err := tx.Where("user_id = ?", userID).First(&scheme).Error; err == nil {
			schemeRate := decimal.NewFromFloat(scheme.MyRate).Round(2)
			if schemeRate.GreaterThan(rate) {
				return schemeRate, nil
			}
		}
		return rate, nil
	}

	var scheme models.AffiliateLevelScheme
	if err := tx.Where("user_id = ?", userID).First(&scheme).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return decimal.Zero, nil
		}
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(scheme.MyRate).Round(2), nil
}

func (s *AffiliateService) getActiveAffiliateProfileByUserID(tx *gorm.DB, userID uint) (*models.AffiliateProfile, error) {
	if tx == nil || userID == 0 {
		return nil, nil
	}
	var profile models.AffiliateProfile
	if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if strings.TrimSpace(profile.Status) != constants.AffiliateProfileStatusActive {
		return nil, nil
	}
	return &profile, nil
}

// GetUserDashboard 获取用户返利中心数据
func (s *AffiliateService) GetUserDashboard(userID uint) (AffiliateDashboard, error) {
	dashboard := AffiliateDashboard{
		Opened:              false,
		PendingCommission:   models.NewMoneyFromDecimal(decimal.Zero),
		AvailableCommission: models.NewMoneyFromDecimal(decimal.Zero),
		WithdrawnCommission: models.NewMoneyFromDecimal(decimal.Zero),
	}
	if userID == 0 || s.repo == nil {
		return dashboard, nil
	}
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return dashboard, err
	}
	if profile == nil {
		return dashboard, nil
	}

	stats, err := s.buildProfileStats(profile.ID)
	if err != nil {
		return dashboard, err
	}
	dashboard.Opened = true
	dashboard.AffiliateCode = profile.AffiliateCode
	dashboard.PromotionPath = "/?aff=" + profile.AffiliateCode
	dashboard.ClickCount = stats.ClickCount
	dashboard.ValidOrderCount = stats.ValidOrderCount
	dashboard.ConversionRate = stats.ConversionRate
	dashboard.PendingCommission = stats.PendingCommission
	dashboard.AvailableCommission = stats.AvailableCommission
	dashboard.WithdrawnCommission = stats.WithdrawnCommission
	return dashboard, nil
}

// ListUserCommissions 查询用户佣金记录
func (s *AffiliateService) ListUserCommissions(userID uint, page, pageSize int, status string) ([]models.AffiliateCommission, int64, error) {
	if userID == 0 || s.repo == nil {
		return []models.AffiliateCommission{}, 0, nil
	}
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if profile == nil {
		return []models.AffiliateCommission{}, 0, nil
	}
	return s.repo.ListCommissions(repository.AffiliateCommissionListFilter{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: profile.ID,
		Status:             strings.TrimSpace(status),
	})
}

// ListUserWithdraws 查询用户提现记录
func (s *AffiliateService) ListUserWithdraws(userID uint, page, pageSize int, status string) ([]models.AffiliateWithdrawRequest, int64, error) {
	if userID == 0 || s.repo == nil {
		return []models.AffiliateWithdrawRequest{}, 0, nil
	}
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if profile == nil {
		return []models.AffiliateWithdrawRequest{}, 0, nil
	}
	return s.repo.ListWithdraws(repository.AffiliateWithdrawListFilter{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: profile.ID,
		Status:             strings.TrimSpace(status),
	})
}

// ApplyWithdraw 用户提交提现申请
func (s *AffiliateService) ApplyWithdraw(userID uint, input AffiliateWithdrawApplyInput) (*models.AffiliateWithdrawRequest, error) {
	if userID == 0 || s.repo == nil {
		return nil, ErrAffiliateNotOpened
	}
	setting, err := s.settingService.GetAffiliateSetting()
	if err != nil {
		return nil, err
	}
	if !setting.Enabled {
		return nil, ErrAffiliateDisabled
	}

	amount := input.Amount.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrAffiliateWithdrawAmountInvalid
	}
	minAmount := decimal.NewFromFloat(setting.MinWithdrawAmount).Round(2)
	if amount.LessThan(minAmount) {
		return nil, ErrAffiliateWithdrawAmountInvalid
	}
	channel := strings.TrimSpace(input.Channel)
	account := strings.TrimSpace(input.Account)
	if channel == "" || account == "" {
		return nil, ErrAffiliateWithdrawChannelInvalid
	}
	if len(setting.WithdrawChannels) > 0 && !containsWithdrawChannel(setting.WithdrawChannels, channel) {
		return nil, ErrAffiliateWithdrawChannelInvalid
	}
	if err := s.ConfirmDueCommissions(time.Now()); err != nil {
		return nil, err
	}

	var createdID uint
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		repoTx := s.repo.WithTx(tx)
		profile, err := repoTx.GetProfileByUserID(userID)
		if err != nil {
			return err
		}
		if profile == nil {
			return ErrAffiliateNotOpened
		}
		if strings.TrimSpace(profile.Status) != constants.AffiliateProfileStatusActive {
			return ErrAffiliateNotOpened
		}

		commissions, err := repoTx.ListAvailableCommissionsForUpdate(profile.ID)
		if err != nil {
			return err
		}

		remaining := amount
		selectedIDs := make([]uint, 0)
		now := time.Now()
		for _, commission := range commissions {
			if remaining.LessThanOrEqual(decimal.Zero) {
				break
			}
			rowAmount := commission.CommissionAmount.Decimal.Round(2)
			if rowAmount.LessThanOrEqual(decimal.Zero) {
				continue
			}
			if rowAmount.LessThanOrEqual(remaining) {
				selectedIDs = append(selectedIDs, commission.ID)
				remaining = remaining.Sub(rowAmount).Round(2)
				continue
			}

			// 最后一条记录金额大于申请剩余金额时，拆分记录避免超额冻结。
			boundAmount := remaining.Round(2)
			remainAmount := rowAmount.Sub(boundAmount).Round(2)
			commission.CommissionAmount = models.NewMoneyFromDecimal(boundAmount)
			commission.UpdatedAt = now
			if err := repoTx.UpdateCommission(&commission); err != nil {
				return err
			}

			remainCommission := commission
			remainCommission.ID = 0
			remainCommission.CommissionType = buildSplitCommissionType(commission.ID)
			remainCommission.CommissionAmount = models.NewMoneyFromDecimal(remainAmount)
			remainCommission.WithdrawRequestID = nil
			remainCommission.Status = constants.AffiliateCommissionStatusAvailable
			remainCommission.InvalidReason = ""
			remainCommission.CreatedAt = now
			remainCommission.UpdatedAt = now
			if err := repoTx.CreateCommission(&remainCommission); err != nil {
				return err
			}

			selectedIDs = append(selectedIDs, commission.ID)
			remaining = decimal.Zero
			break
		}
		if remaining.GreaterThan(decimal.Zero) {
			return ErrAffiliateWithdrawInsufficient
		}

		req := &models.AffiliateWithdrawRequest{
			AffiliateProfileID: profile.ID,
			Amount:             models.NewMoneyFromDecimal(amount),
			Channel:            channel,
			Account:            account,
			Status:             constants.AffiliateWithdrawStatusPendingReview,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := repoTx.CreateWithdraw(req); err != nil {
			return err
		}
		if err := repoTx.BatchUpdateCommissions(selectedIDs, map[string]interface{}{
			"withdraw_request_id": req.ID,
			"updated_at":          now,
		}); err != nil {
			return err
		}
		createdID = req.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repo.GetWithdrawByID(createdID)
}

// ReviewWithdraw 管理端审核提现申请
func (s *AffiliateService) ReviewWithdraw(adminID, withdrawID uint, action, rejectReason string) (*models.AffiliateWithdrawRequest, error) {
	if withdrawID == 0 || s.repo == nil {
		return nil, ErrNotFound
	}
	act := strings.ToLower(strings.TrimSpace(action))
	if act != constants.AffiliateWithdrawActionReject && act != constants.AffiliateWithdrawActionPay {
		return nil, ErrAffiliateWithdrawStatusInvalid
	}
	rejectReason = strings.TrimSpace(rejectReason)

	err := s.repo.Transaction(func(tx *gorm.DB) error {
		repoTx := s.repo.WithTx(tx)
		req, err := repoTx.GetWithdrawByIDForUpdate(withdrawID)
		if err != nil {
			return err
		}
		if req == nil {
			return ErrNotFound
		}
		if req.Status != constants.AffiliateWithdrawStatusPendingReview {
			return ErrAffiliateWithdrawStatusInvalid
		}

		commissions, err := repoTx.ListCommissionsByWithdrawIDForUpdate(withdrawID)
		if err != nil {
			return err
		}
		ids := make([]uint, 0, len(commissions))
		for _, commission := range commissions {
			ids = append(ids, commission.ID)
		}

		now := time.Now()
		req.ProcessedBy = &adminID
		req.ProcessedAt = &now
		req.UpdatedAt = now
		if act == constants.AffiliateWithdrawActionReject {
			req.Status = constants.AffiliateWithdrawStatusRejected
			req.RejectReason = rejectReason
			if err := repoTx.BatchUpdateCommissions(ids, map[string]interface{}{
				"withdraw_request_id": nil,
				"updated_at":          now,
			}); err != nil {
				return err
			}
		} else {
			req.Status = constants.AffiliateWithdrawStatusPaid
			req.RejectReason = ""
			if err := repoTx.BatchUpdateCommissions(ids, map[string]interface{}{
				"status":     constants.AffiliateCommissionStatusWithdrawn,
				"updated_at": now,
			}); err != nil {
				return err
			}
		}
		return repoTx.UpdateWithdraw(req)
	})
	if err != nil {
		return nil, err
	}
	return s.repo.GetWithdrawByID(withdrawID)
}

// ListAdminUsers 后台查询推广用户列表
func (s *AffiliateService) ListAdminUsers(filter repository.AffiliateProfileListFilter) ([]AffiliateAdminUserItem, int64, error) {
	if s.repo == nil {
		return []AffiliateAdminUserItem{}, 0, nil
	}
	rows, total, err := s.repo.ListProfiles(filter)
	if err != nil {
		return nil, 0, err
	}
	profileIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.ID == 0 {
			continue
		}
		profileIDs = append(profileIDs, row.ID)
	}
	statsMap, err := s.repo.GetProfileStatsBatch(profileIDs)
	if err != nil {
		return nil, 0, err
	}

	topDiscountRateMap := map[uint]float64{}
	hasParentPromoterMap := map[uint]bool{}
	if models.DB != nil && len(rows) > 0 {
		userIDs := make([]uint, 0, len(rows))
		for _, row := range rows {
			if row.UserID > 0 {
				userIDs = append(userIDs, row.UserID)
			}
		}
		if len(userIDs) > 0 {
			var schemes []models.AffiliateLevelScheme
			if err := models.DB.Select("user_id", "my_rate").Where("user_id IN ?", userIDs).Find(&schemes).Error; err != nil {
				return nil, 0, err
			}
			for _, scheme := range schemes {
				topDiscountRateMap[scheme.UserID] = scheme.MyRate
			}

			var promotionLevels []models.UserPromotionLevel
			if err := models.DB.Select("user_id", "parent_user_id").Where("user_id IN ?", userIDs).Find(&promotionLevels).Error; err != nil {
				return nil, 0, err
			}
			for _, level := range promotionLevels {
				hasParentPromoterMap[level.UserID] = level.ParentUserID > 0
			}
		}
	}
	result := make([]AffiliateAdminUserItem, 0, len(rows))
	for _, row := range rows {
		agg := statsMap[row.ID]
		row.IsTokenMerchant = row.User.IsTokenMerchant
		row.TopDiscountRate = topDiscountRateMap[row.UserID]
		stats := AffiliateStats{
			ClickCount:          agg.ClickCount,
			ValidOrderCount:     agg.ValidOrderCount,
			ConversionRate:      calcAffiliateConversion(agg.ValidOrderCount, agg.ClickCount),
			PendingCommission:   models.NewMoneyFromDecimal(agg.PendingCommission.Round(2)),
			AvailableCommission: models.NewMoneyFromDecimal(agg.AvailableCommission.Round(2)),
			WithdrawnCommission: models.NewMoneyFromDecimal(agg.WithdrawnCommission.Round(2)),
		}
		result = append(result, AffiliateAdminUserItem{
			Profile:           row,
			Stats:             stats,
			TopDiscountRate:   topDiscountRateMap[row.UserID],
			HasParentPromoter: hasParentPromoterMap[row.UserID],
		})
	}
	return result, total, nil
}

// AdminAuthorizeTokenMerchant 管理端手动授权 Token 商身份。
func (s *AffiliateService) AdminAuthorizeTokenMerchant(profileID uint) (*models.AffiliateProfile, error) {
	if profileID == 0 || s.repo == nil {
		return nil, ErrNotFound
	}
	profile, err := s.repo.GetProfileByID(profileID)
	if err != nil {
		return nil, err
	}
	if profile == nil || profile.UserID == 0 {
		return nil, ErrNotFound
	}
	return s.OpenTokenMerchant(profile.UserID, "")
}

// ListAdminCommissions 后台查询佣金记录
func (s *AffiliateService) ListAdminCommissions(filter AffiliateAdminCommissionListFilter) ([]models.AffiliateCommission, int64, error) {
	if s.repo == nil {
		return []models.AffiliateCommission{}, 0, nil
	}
	return s.repo.ListCommissions(repository.AffiliateCommissionListFilter{
		Page:               filter.Page,
		PageSize:           filter.PageSize,
		AffiliateProfileID: filter.AffiliateProfileID,
		OrderNo:            strings.TrimSpace(filter.OrderNo),
		Status:             strings.TrimSpace(filter.Status),
		Keyword:            strings.TrimSpace(filter.Keyword),
	})
}

// ListAdminWithdraws 后台查询提现申请
func (s *AffiliateService) ListAdminWithdraws(filter AffiliateAdminWithdrawListFilter) ([]models.AffiliateWithdrawRequest, int64, error) {
	if s.repo == nil {
		return []models.AffiliateWithdrawRequest{}, 0, nil
	}
	return s.repo.ListWithdraws(repository.AffiliateWithdrawListFilter{
		Page:               filter.Page,
		PageSize:           filter.PageSize,
		AffiliateProfileID: filter.AffiliateProfileID,
		Status:             strings.TrimSpace(filter.Status),
		Keyword:            strings.TrimSpace(filter.Keyword),
	})
}

func (s *AffiliateService) resolveAffiliateProfileForOrder(order *models.Order) (*models.AffiliateProfile, error) {
	if order == nil || s.repo == nil {
		return nil, nil
	}
	if order.AffiliateProfileID != nil && *order.AffiliateProfileID > 0 {
		return s.repo.GetProfileByID(*order.AffiliateProfileID)
	}
	if strings.TrimSpace(order.AffiliateCode) != "" {
		return s.repo.GetProfileByCode(order.AffiliateCode)
	}
	if order.UserID != 0 && models.DB != nil {
		return s.resolveAffiliateProfileByPromotionRelation(models.DB, order.UserID)
	}
	return nil, nil
}

func (s *AffiliateService) resolveAffiliateProfileByPromotionRelation(tx *gorm.DB, userID uint) (*models.AffiliateProfile, error) {
	if tx == nil || userID == 0 {
		return nil, nil
	}
	var level models.UserPromotionLevel
	if err := tx.Select("parent_user_id").Where("user_id = ?", userID).First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if level.ParentUserID == 0 {
		return nil, nil
	}
	return s.getActiveAffiliateProfileByUserID(tx, level.ParentUserID)
}

func (s *AffiliateService) syncOrderAffiliateSnapshotTx(tx *gorm.DB, order *models.Order, profile *models.AffiliateProfile) error {
	if tx == nil || order == nil || order.ID == 0 || profile == nil || profile.ID == 0 {
		return nil
	}
	needProfileID := order.AffiliateProfileID == nil || *order.AffiliateProfileID == 0
	needCode := strings.TrimSpace(order.AffiliateCode) == ""
	if !needProfileID && !needCode {
		return nil
	}
	updates := map[string]interface{}{}
	if needProfileID {
		updates["affiliate_profile_id"] = profile.ID
		order.AffiliateProfileID = &profile.ID
	}
	if needCode {
		updates["affiliate_code"] = profile.AffiliateCode
		order.AffiliateCode = profile.AffiliateCode
	}
	if len(updates) == 0 {
		return nil
	}
	query := tx.Model(&models.Order{}).Where("id = ?", order.ID)
	if order.ParentID == nil {
		query = query.Or("parent_id = ?", order.ID)
	}
	return query.Updates(updates).Error
}

func (s *AffiliateService) isTokenMerchantUser(userID uint) bool {
	if userID == 0 {
		return false
	}
	if s.userRepo != nil {
		user, err := s.userRepo.GetByID(userID)
		if err == nil && user != nil {
			return user.IsTokenMerchant
		}
	}
	if models.DB == nil {
		return false
	}
	var user models.User
	if err := models.DB.Select("id", "is_token_merchant").First(&user, userID).Error; err != nil {
		return false
	}
	return user.IsTokenMerchant
}

func (s *AffiliateService) calculateCommissionBaseAmount(order *models.Order) (decimal.Decimal, error) {
	if order == nil || s.productRepo == nil {
		return decimal.Zero, nil
	}
	productIDs := collectAffiliateProductIDs(order)
	if len(productIDs) == 0 {
		return decimal.Zero, nil
	}
	products, err := s.productRepo.ListByIDs(productIDs)
	if err != nil {
		return decimal.Zero, err
	}
	productMap := make(map[uint]models.Product, len(products))
	for _, product := range products {
		productMap[product.ID] = product
	}

	targetOrders := order.Children
	if len(targetOrders) == 0 {
		targetOrders = []models.Order{*order}
	}

	total := decimal.Zero
	for _, current := range targetOrders {
		for _, item := range current.Items {
			product, ok := productMap[item.ProductID]
			if !ok || !product.IsAffiliateEnabled {
				continue
			}
			payable := item.TotalPrice.Decimal.Sub(item.CouponDiscount.Decimal).Round(2)
			if payable.LessThan(decimal.Zero) {
				payable = decimal.Zero
			}
			total = total.Add(payable).Round(2)
		}
	}
	return total, nil
}

func collectAffiliateProductIDs(order *models.Order) []uint {
	if order == nil {
		return nil
	}
	ids := make([]uint, 0)
	seen := make(map[uint]struct{})
	appendItem := func(item models.OrderItem) {
		if item.ProductID == 0 {
			return
		}
		if _, ok := seen[item.ProductID]; ok {
			return
		}
		seen[item.ProductID] = struct{}{}
		ids = append(ids, item.ProductID)
	}
	for _, item := range order.Items {
		appendItem(item)
	}
	for _, child := range order.Children {
		for _, item := range child.Items {
			appendItem(item)
		}
	}
	return ids
}

func (s *AffiliateService) buildProfileStats(profileID uint) (AffiliateStats, error) {
	stats := AffiliateStats{
		PendingCommission:   models.NewMoneyFromDecimal(decimal.Zero),
		AvailableCommission: models.NewMoneyFromDecimal(decimal.Zero),
		WithdrawnCommission: models.NewMoneyFromDecimal(decimal.Zero),
	}
	if profileID == 0 || s.repo == nil {
		return stats, nil
	}
	clickCount, err := s.repo.CountClicksByProfile(profileID)
	if err != nil {
		return stats, err
	}
	validOrders, err := s.repo.CountValidOrdersByProfile(profileID)
	if err != nil {
		return stats, err
	}
	pendingAmount, err := s.repo.SumCommissionByProfile(profileID, []string{
		constants.AffiliateCommissionStatusPendingConfirm,
	}, false)
	if err != nil {
		return stats, err
	}
	availableAmount, err := s.repo.SumCommissionByProfile(profileID, []string{
		constants.AffiliateCommissionStatusAvailable,
	}, true)
	if err != nil {
		return stats, err
	}
	withdrawnAmount, err := s.repo.SumCommissionByProfile(profileID, []string{
		constants.AffiliateCommissionStatusWithdrawn,
	}, false)
	if err != nil {
		return stats, err
	}

	stats.ClickCount = clickCount
	stats.ValidOrderCount = validOrders
	stats.ConversionRate = calcAffiliateConversion(validOrders, clickCount)
	stats.PendingCommission = models.NewMoneyFromDecimal(pendingAmount)
	stats.AvailableCommission = models.NewMoneyFromDecimal(availableAmount)
	stats.WithdrawnCommission = models.NewMoneyFromDecimal(withdrawnAmount)
	return stats, nil
}

func calcAffiliateConversion(validOrders, clicks int64) float64 {
	if clicks <= 0 || validOrders <= 0 {
		return 0
	}
	value := (float64(validOrders) / float64(clicks)) * 100
	return math.Round(value*100) / 100
}

func containsWithdrawChannel(channels []string, channel string) bool {
	target := strings.ToLower(strings.TrimSpace(channel))
	if target == "" {
		return false
	}
	for _, item := range channels {
		if strings.ToLower(strings.TrimSpace(item)) == target {
			return true
		}
	}
	return false
}

func normalizeAffiliateProfileIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return []uint{}
	}
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func generateAffiliateCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var builder strings.Builder
	builder.Grow(affiliateCodeLength)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < affiliateCodeLength; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		builder.WriteByte(alphabet[n.Int64()])
	}
	return builder.String(), nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}

func buildSplitCommissionType(sourceID uint) string {
	suffix := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
	base := affiliateSplitTypePrefix + strconv.FormatUint(uint64(sourceID), 36)
	result := base + suffix
	if len(result) > 20 {
		return result[:20]
	}
	return result
}
