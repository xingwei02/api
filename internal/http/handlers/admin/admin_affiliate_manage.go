package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	"gorm.io/gorm"
)

// AffiliateProfileStatusRequest 返利用户状态更新请求
type AffiliateProfileStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// BatchAffiliateProfileStatusRequest 返利用户批量状态更新请求
type BatchAffiliateProfileStatusRequest struct {
	ProfileIDs []uint `json:"profile_ids" binding:"required"`
	Status     string `json:"status" binding:"required"`
}

// AdminAffiliateDiscountRequest 管理端 Token 商优惠配置
type AdminAffiliateDiscountRequest struct {
	DiscountRate        float64 `json:"discount_rate"`
	MerchantPageEnabled bool    `json:"merchant_page_enabled"`
	GroupSectionEnabled bool    `json:"group_section_enabled"`
}

type adminAffiliateDiscountResponse struct {
	UserID              uint    `json:"user_id"`
	DiscountRate        float64 `json:"discount_rate"`
	MerchantPageEnabled bool    `json:"merchant_page_enabled"`
	GroupSectionEnabled bool    `json:"group_section_enabled"`
}

// AdminAffiliateContactRequest 管理端 Token 商官方群内容配置
type AdminAffiliateContactRequest struct {
	Phone               string `json:"phone"`
	Notice              string `json:"notice"`
	GroupImageURL       string `json:"group_image_url"`
	ParentGroupImageURL string `json:"parent_group_image_url"`
}

// ListAffiliateUsers 管理端推广用户列表
func (h *Handler) ListAffiliateUsers(c *gin.Context) {
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = shared.NormalizePagination(page, pageSize)
	userID, _ := shared.ParseQueryUint(c.Query("user_id"), false)

	rows, total, err := h.AffiliateService.ListAdminUsers(repository.AffiliateProfileListFilter{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
		Status:   strings.TrimSpace(c.Query("status")),
		Code:     strings.TrimSpace(c.Query("code")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// ListAffiliateCommissions 管理端佣金列表
func (h *Handler) ListAffiliateCommissions(c *gin.Context) {
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = shared.NormalizePagination(page, pageSize)
	profileID, _ := shared.ParseQueryUint(c.Query("affiliate_profile_id"), false)

	rows, total, err := h.AffiliateService.ListAdminCommissions(service.AffiliateAdminCommissionListFilter{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: profileID,
		OrderNo:            strings.TrimSpace(c.Query("order_no")),
		Status:             strings.TrimSpace(c.Query("status")),
		Keyword:            strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// ListAffiliateWithdraws 管理端提现审核列表
func (h *Handler) ListAffiliateWithdraws(c *gin.Context) {
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = shared.NormalizePagination(page, pageSize)
	profileID, _ := shared.ParseQueryUint(c.Query("affiliate_profile_id"), false)

	rows, total, err := h.AffiliateService.ListAdminWithdraws(service.AffiliateAdminWithdrawListFilter{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: profileID,
		Status:             strings.TrimSpace(c.Query("status")),
		Keyword:            strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// UpdateAffiliateUserStatus 管理端更新返利用户状态
func (h *Handler) UpdateAffiliateUserStatus(c *gin.Context) {
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", nil)
		return
	}
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var req AffiliateProfileStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	row, err := h.AffiliateService.UpdateAffiliateProfileStatus(id, strings.TrimSpace(req.Status))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
		case errors.Is(err, service.ErrAffiliateProfileStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, row)
}

// BatchUpdateAffiliateUserStatus 管理端批量更新返利用户状态
func (h *Handler) BatchUpdateAffiliateUserStatus(c *gin.Context) {
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", nil)
		return
	}
	var req BatchAffiliateProfileStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if len(req.ProfileIDs) == 0 {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	updated, err := h.AffiliateService.BatchUpdateAffiliateProfileStatus(req.ProfileIDs, strings.TrimSpace(req.Status))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAffiliateProfileStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, gin.H{"updated": updated})
}

// GetAffiliateUserDiscount 管理端获取 Token 商折扣配置
func (h *Handler) GetAffiliateUserDiscount(c *gin.Context) {
	profileID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var profile models.AffiliateProfile
	if err := models.DB.Select("id", "user_id").First(&profile, profileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}

	resp := adminAffiliateDiscountResponse{UserID: profile.UserID}
	var discount models.AffiliateDiscount
	if err := models.DB.Where("user_id = ?", profile.UserID).First(&discount).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
			return
		}
	} else {
		resp.DiscountRate = discount.DiscountRate
		resp.MerchantPageEnabled = discount.MerchantPageEnabled
		resp.GroupSectionEnabled = discount.GroupSectionEnabled
	}
	response.Success(c, resp)
}

// UpdateAffiliateUserDiscount 管理端更新 Token 商折扣配置
func (h *Handler) UpdateAffiliateUserDiscount(c *gin.Context) {
	profileID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var req AdminAffiliateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if req.DiscountRate < 0 || req.DiscountRate > 5 {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var profile models.AffiliateProfile
	if err := models.DB.Select("id", "user_id").First(&profile, profileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}

	var discount models.AffiliateDiscount
	result := models.DB.Where("user_id = ?", profile.UserID).First(&discount)
	discount.UserID = profile.UserID
	discount.DiscountRate = req.DiscountRate
	discount.MerchantPageEnabled = req.MerchantPageEnabled
	discount.GroupSectionEnabled = req.GroupSectionEnabled
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			shared.RespondError(c, response.CodeInternal, "error.save_failed", result.Error)
			return
		}
		if err := models.DB.Create(&discount).Error; err != nil {
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
			return
		}
	} else {
		if err := models.DB.Save(&discount).Error; err != nil {
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
			return
		}
	}
	response.Success(c, adminAffiliateDiscountResponse{
		UserID:              profile.UserID,
		DiscountRate:        discount.DiscountRate,
		MerchantPageEnabled: discount.MerchantPageEnabled,
		GroupSectionEnabled: discount.GroupSectionEnabled,
	})
}

// AuthorizeAffiliateTokenMerchant 管理端手动授权 Token 商。
func (h *Handler) AuthorizeAffiliateTokenMerchant(c *gin.Context) {
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", nil)
		return
	}
	profileID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	profile, err := h.AffiliateService.AdminAuthorizeTokenMerchant(profileID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
		case errors.Is(err, service.ErrAffiliateDisabled):
			shared.RespondError(c, response.CodeBadRequest, "error.forbidden", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, profile)
}

// GetAffiliateUserContact 管理端获取 Token 商官方群内容配置
func (h *Handler) GetAffiliateUserContact(c *gin.Context) {
	profileID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var profile models.AffiliateProfile
	if err := models.DB.Select("id", "user_id").First(&profile, profileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}

	contact := models.AffiliateContact{UserID: profile.UserID}
	if err := models.DB.Where("user_id = ?", profile.UserID).First(&contact).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
			return
		}
	}
	response.Success(c, contact)
}

// UpdateAffiliateUserContact 管理端更新 Token 商官方群内容配置
func (h *Handler) UpdateAffiliateUserContact(c *gin.Context) {
	profileID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var req AdminAffiliateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	var profile models.AffiliateProfile
	if err := models.DB.Select("id", "user_id").First(&profile, profileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}

	var contact models.AffiliateContact
	result := models.DB.Where("user_id = ?", profile.UserID).First(&contact)
	contact.UserID = profile.UserID
	contact.Phone = strings.TrimSpace(req.Phone)
	contact.Notice = strings.TrimSpace(req.Notice)
	contact.GroupImageURL = strings.TrimSpace(req.GroupImageURL)
	contact.ParentGroupImageURL = strings.TrimSpace(req.ParentGroupImageURL)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			shared.RespondError(c, response.CodeInternal, "error.save_failed", result.Error)
			return
		}
		if err := models.DB.Create(&contact).Error; err != nil {
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
			return
		}
	} else {
		if err := models.DB.Save(&contact).Error; err != nil {
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
			return
		}
	}
	response.Success(c, contact)
}

// AffiliateReviewWithdrawRequest 提现审核请求
type AffiliateReviewWithdrawRequest struct {
	Reason string `json:"reason"`
}

// RejectAffiliateWithdraw 拒绝提现申请
func (h *Handler) RejectAffiliateWithdraw(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", nil)
		return
	}
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var req AffiliateReviewWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	row, err := h.AffiliateService.ReviewWithdraw(adminID, id, constants.AffiliateWithdrawActionReject, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
		case errors.Is(err, service.ErrAffiliateWithdrawStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, row)
}

// PayAffiliateWithdraw 标记提现已支付
func (h *Handler) PayAffiliateWithdraw(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}
	if h.AffiliateService == nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", nil)
		return
	}
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	row, err := h.AffiliateService.ReviewWithdraw(adminID, id, constants.AffiliateWithdrawActionPay, "")
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
		case errors.Is(err, service.ErrAffiliateWithdrawStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, row)
}
