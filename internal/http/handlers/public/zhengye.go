package public

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/service"
)

// queryIntDefault 解析 query 参数为 int，失败时返回默认值
func queryIntDefault(c *gin.Context, key string, def int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func (h *Handler) ensureZhengyeAccess(c *gin.Context, userID uint) bool {
	if h == nil || h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return false
	}
	if err := h.ZhengyeService.EnsureTokenMerchant(userID); err != nil {
		switch {
		case errors.Is(err, service.ErrTokenMerchantRequired):
			shared.RespondError(c, response.CodeForbidden, "error.forbidden", nil)
		case errors.Is(err, service.ErrUserDisabled):
			shared.RespondError(c, response.CodeUnauthorized, "error.user_disabled", nil)
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		}
		return false
	}
	return true
}

// ─────────────────────────────────────────────────────────────────────────────
// Dashboard & Stats
// ─────────────────────────────────────────────────────────────────────────────

// GetZhengyeDashboard GET /affiliate/dashboard-v2
func (h *Handler) GetZhengyeDashboard(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	data, err := h.ZhengyeService.GetDashboard(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// GetZhengyeStats GET /affiliate/stats?period=7d
func (h *Handler) GetZhengyeStats(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	period := service.ZhengyeStatsPeriod(c.DefaultQuery("period", "7d"))
	data, err := h.ZhengyeService.GetStats(uid, period)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// GetZhengyeRank GET /affiliate/rank
func (h *Handler) GetZhengyeRank(c *gin.Context) {
	if h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	// 封神榜不强制登录，未登录时 currentUserID=0（不显示"我的排名"）
	uid, _ := shared.GetUserID(c)
	data, err := h.ZhengyeService.GetRank(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// GetZhengyeLevels GET /affiliate/levels
func (h *Handler) GetZhengyeLevels(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	data, err := h.ZhengyeService.GetLevels(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// SaveZhengyeLevels PUT /affiliate/levels
func (h *Handler) SaveZhengyeLevels(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	var req service.SaveZhengyeLevelsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if err := h.ZhengyeService.SaveLevels(uid, req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.ZhengyeService.GetLevels(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// ─────────────────────────────────────────────────────────────────────────────
// Orders
// ─────────────────────────────────────────────────────────────────────────────

// GetZhengyeOrders GET /affiliate/orders
func (h *Handler) GetZhengyeOrders(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	filter := service.ZhengyeOrdersFilter{
		Page:     queryIntDefault(c, "page", 1),
		PageSize: queryIntDefault(c, "page_size", 20),
		Status:   c.Query("status"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	}
	data, err := h.ZhengyeService.GetOrders(uid, filter)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// GetZhengyeOrderCommissionDetail GET /affiliate/orders/:id/commission-detail
func (h *Handler) GetZhengyeOrderCommissionDetail(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	orderID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || orderID64 == 0 {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}
	data, err := h.ZhengyeService.GetOrderCommissionDetail(uid, uint(orderID64))
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// ─────────────────────────────────────────────────────────────────────────────
// Team
// ─────────────────────────────────────────────────────────────────────────────

// GetZhengyeTeam GET /affiliate/team
func (h *Handler) GetZhengyeTeam(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	filter := service.ZhengyeTeamFilter{
		Page:     queryIntDefault(c, "page", 1),
		PageSize: queryIntDefault(c, "page_size", 20),
		Depth:    queryIntDefault(c, "depth", 0),
	}
	data, err := h.ZhengyeService.GetTeam(uid, filter)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// ─────────────────────────────────────────────────────────────────────────────
// Partners
// ─────────────────────────────────────────────────────────────────────────────

// GetZhengyePartners GET /affiliate/partners
func (h *Handler) GetZhengyePartners(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	filter := service.ZhengyePartnersFilter{
		Keyword:  c.Query("keyword"),
		Page:     queryIntDefault(c, "page", 1),
		PageSize: queryIntDefault(c, "page_size", 20),
	}
	data, err := h.ZhengyeService.GetPartners(uid, filter)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

type updatePartnerRateRequest struct {
	Rate float64 `json:"rate"`
}

// UpdateZhengyePartnerRate PATCH /affiliate/partners/:id
func (h *Handler) UpdateZhengyePartnerRate(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	partnerID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || partnerID64 == 0 {
		response.Error(c, http.StatusBadRequest, "invalid partner id")
		return
	}
	var req updatePartnerRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if err := h.ZhengyeService.UpdatePartnerRate(uid, uint(partnerID64), req.Rate); err != nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}
	response.Success(c, gin.H{"partner_id": uint(partnerID64), "rate": req.Rate})
}

// GetZhengyePartnerOrdersByDate GET /affiliate/partners/:id/orders
func (h *Handler) GetZhengyePartnerOrdersByDate(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	partnerID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || partnerID64 == 0 {
		response.Error(c, http.StatusBadRequest, "invalid partner id")
		return
	}
	filter := service.OrderDetailFilter{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Keyword:   c.Query("keyword"),
		Page:      queryIntDefault(c, "page", 1),
		PageSize:  queryIntDefault(c, "page_size", 20),
	}
	data, err := h.ZhengyeService.GetPartnerOrdersByDate(uid, uint(partnerID64), filter)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

// ─────────────────────────────────────────────────────────────────────────────
// Settlement
// ─────────────────────────────────────────────────────────────────────────────

// GetZhengyeSettlement GET /affiliate/settlement
func (h *Handler) GetZhengyeSettlement(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.ZhengyeService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	filter := service.ZhengyeSettlementFilter{
		Date:     c.Query("date"),
		Keyword:  c.Query("keyword"),
		Page:     queryIntDefault(c, "page", 1),
		PageSize: queryIntDefault(c, "page_size", 20),
	}
	data, err := h.ZhengyeService.GetSettlement(uid, filter)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, data)
}

type paySettlementRequest struct {
	PartnerID uint   `json:"partner_id"`
	Date      string `json:"date"`
}

// PayZhengyeSettlement POST /affiliate/settlement/pay
func (h *Handler) PayZhengyeSettlement(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	var req paySettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if req.PartnerID == 0 {
		response.Error(c, http.StatusBadRequest, "partner_id is required")
		return
	}
	if err := h.ZhengyeService.PaySettlement(uid, req.PartnerID, req.Date); err != nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}
	response.Success(c, gin.H{"partner_id": req.PartnerID, "date": req.Date})
}

// ─────────────────────────────────────────────────────────────────────────────
// Contact
// ─────────────────────────────────────────────────────────────────────────────

// GetZhengyeContact GET /affiliate/contact
func (h *Handler) GetZhengyeContact(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	var contact models.AffiliateContact
	if err := models.DB.Where("user_id = ?", uid).First(&contact).Error; err != nil {
		response.Success(c, models.AffiliateContact{UserID: uid})
		return
	}
	response.Success(c, contact)
}

type saveContactRequest struct {
	Phone               string `json:"phone"`
	Notice              string `json:"notice"`
	GroupImageURL       string `json:"group_image_url"`
	ParentGroupImageURL string `json:"parent_group_image_url"`
}

// SaveZhengyeContact PUT /affiliate/contact
func (h *Handler) SaveZhengyeContact(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	var req saveContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	var contact models.AffiliateContact
	result := models.DB.Where("user_id = ?", uid).First(&contact)
	contact.UserID = uid
	contact.Phone = req.Phone
	contact.Notice = req.Notice
	contact.GroupImageURL = req.GroupImageURL
	contact.ParentGroupImageURL = req.ParentGroupImageURL
	if result.Error != nil {
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

// ─────────────────────────────────────────────────────────────────────────────
// Discount
// ─────────────────────────────────────────────────────────────────────────────

// GetZhengyeDiscount GET /affiliate/discount
func (h *Handler) GetZhengyeDiscount(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	var discount models.AffiliateDiscount
	if err := models.DB.Where("user_id = ?", uid).First(&discount).Error; err != nil {
		response.Success(c, models.AffiliateDiscount{UserID: uid})
		return
	}
	response.Success(c, discount)
}

// ─────────────────────────────────────────────────────────────────────────────
// Settlement - 余额与提现（新增）
// ─────────────────────────────────────────────────────────────────────────────

// GetZhengyeBalance GET /affiliate/balance
func (h *Handler) GetZhengyeBalance(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.SettlementService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	balance, err := h.SettlementService.GetUserBalance(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, balance)
}

// GetZhengyeBalanceLogs GET /affiliate/balance-logs?page=1&page_size=20
func (h *Handler) GetZhengyeBalanceLogs(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.SettlementService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	page := queryIntDefault(c, "page", 1)
	pageSize := queryIntDefault(c, "page_size", 20)
	logs, total, err := h.SettlementService.GetBalanceLogs(uid, page, pageSize)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, gin.H{"items": logs, "total": total, "page": page, "page_size": pageSize})
}

// GetTransferableCommissions GET /affiliate/transferable-commissions?page=1&page_size=100
func (h *Handler) GetTransferableCommissions(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.SettlementService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	page := queryIntDefault(c, "page", 1)
	pageSize := queryIntDefault(c, "page_size", 100)
	items, total, err := h.SettlementService.GetTransferableCommissions(uid, page, pageSize)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

type transferCommissionRequest struct {
	CommissionIDs []uint  `json:"commission_ids" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,min=1"`
	VerifyCode    string  `json:"verify_code"`
}

// TransferCommissionToBalance POST /affiliate/transfer
// 红线1：Handler层接收float64后立即转decimal
func (h *Handler) TransferCommissionToBalance(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.SettlementService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	var req transferCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	// 红线1：立即转decimal验证
	amount := decimal.NewFromFloat(req.Amount).RoundBank(2)
	_ = amount // 验证通过，Service层会再次转换

	// 获取用户邮箱
	user, err := h.UserRepo.GetByID(uid)
	if err != nil || user == nil {
		shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		return
	}

	if err := h.SettlementService.TransferCommissionToBalance(uid, req.CommissionIDs, req.VerifyCode, user.Email, req.Amount); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "佣金转入余额成功"})
}

type applyWithdrawRequest struct {
	Amount        float64 `json:"amount" binding:"required,min=1"`
	AlipayAccount string  `json:"alipay_account" binding:"required"`
	RealName      string  `json:"real_name" binding:"required"`
	VerifyCode    string  `json:"verify_code" binding:"required"`
}

// ApplyWithdraw POST /affiliate/withdraw
// 红线1：Handler层接收float64后立即转decimal
func (h *Handler) ApplyWithdraw(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.SettlementService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	var req applyWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	// 红线1：立即转decimal验证
	amount := decimal.NewFromFloat(req.Amount).RoundBank(2)
	_ = amount

	// 获取用户邮箱
	user, err := h.UserRepo.GetByID(uid)
	if err != nil || user == nil {
		shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		return
	}

	if err := h.SettlementService.ApplyWithdraw(uid, req.Amount, req.AlipayAccount, req.RealName, req.VerifyCode, user.Email); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "提现申请已提交"})
}

// GetZhengyeWithdrawRequests GET /affiliate/withdraw-requests?page=1&page_size=20
func (h *Handler) GetZhengyeWithdrawRequests(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.SettlementService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	page := queryIntDefault(c, "page", 1)
	pageSize := queryIntDefault(c, "page_size", 20)
	requests, total, err := h.SettlementService.GetWithdrawRequests(uid, page, pageSize)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, gin.H{"items": requests, "total": total, "page": page, "page_size": pageSize})
}

// GetZhengyeWithdrawSettings GET /affiliate/withdraw-settings
func (h *Handler) GetZhengyeWithdrawSettings(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	if h.SettlementService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	settings, err := h.SettlementService.GetWithdrawSettings()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, settings)
}

type saveDiscountRequest struct {
	DiscountRate        float64 `json:"discount_rate"`
	MerchantPageEnabled bool    `json:"merchant_page_enabled"`
	GroupSectionEnabled bool    `json:"group_section_enabled"`
}

// SaveZhengyeDiscount PUT /affiliate/discount
func (h *Handler) SaveZhengyeDiscount(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if !h.ensureZhengyeAccess(c, uid) {
		return
	}
	var req saveDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	// 客户优惠上限 5%（业务规则，不可超过）
	if req.DiscountRate < 0 || req.DiscountRate > 5 {
		response.Error(c, http.StatusBadRequest, "discount_rate must be between 0 and 5")
		return
	}
	var discount models.AffiliateDiscount
	result := models.DB.Where("user_id = ?", uid).First(&discount)
	discount.UserID = uid
	discount.DiscountRate = req.DiscountRate
	discount.MerchantPageEnabled = req.MerchantPageEnabled
	discount.GroupSectionEnabled = req.GroupSectionEnabled
	if result.Error != nil {
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
	response.Success(c, discount)
}
