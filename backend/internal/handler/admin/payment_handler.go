package admin

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler handles admin payment management.
type PaymentHandler struct {
	paymentService *service.PaymentService
	configService  *service.PaymentConfigService
}

// NewPaymentHandler creates a new admin PaymentHandler.
func NewPaymentHandler(paymentService *service.PaymentService, configService *service.PaymentConfigService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		configService:  configService,
	}
}

// --- Dashboard ---

// GetDashboard returns payment dashboard statistics.
// GET /api/v1/admin/payment/dashboard
func (h *PaymentHandler) GetDashboard(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	stats, err := h.paymentService.GetDashboardStats(c.Request.Context(), days)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stats)
}

// --- Orders ---

// ListOrders returns a paginated list of all payment orders.
// GET /api/v1/admin/payment/orders
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	var userID int64
	if uid := c.Query("user_id"); uid != "" {
		if v, err := strconv.ParseInt(uid, 10, 64); err == nil {
			userID = v
		}
	}
	orders, total, err := h.paymentService.AdminListOrders(c.Request.Context(), userID, service.OrderListParams{
		Page:        page,
		PageSize:    pageSize,
		Status:      c.Query("status"),
		OrderType:   c.Query("order_type"),
		PaymentType: c.Query("payment_type"),
		Keyword:     c.Query("keyword"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	userIDs := make([]int64, 0, len(orders))
	for _, order := range orders {
		if order != nil {
			userIDs = append(userIDs, order.UserID)
		}
	}
	promoInfo, _ := h.paymentService.GetRegistrationPromoInfo(c.Request.Context(), userIDs)
	response.Paginated(c, adminPaymentOrderResponses(orders, promoInfo), int64(total), page, pageSize)
}

// ExportOrders exports successfully paid orders as an Excel-friendly UTF-8 CSV.
// GET /api/v1/admin/payment/orders/export?start_date=...&end_date=...
func (h *PaymentHandler) ExportOrders(c *gin.Context) {
	loc := exportLocation(c.Query("timezone"))
	start, end, err := parseExportDateRange(c.Query("start_date"), c.Query("end_date"), loc)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	orders, err := h.paymentService.AdminExportPaidOrders(c.Request.Context(), start, end)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	userIDs := make([]int64, 0, len(orders))
	for _, order := range orders {
		if order != nil {
			userIDs = append(userIDs, order.UserID)
		}
	}
	promoInfo, _ := h.paymentService.GetRegistrationPromoInfo(c.Request.Context(), userIDs)

	var builder strings.Builder
	builder.WriteString("\ufeff")
	w := csv.NewWriter(&builder)
	_ = w.Write([]string{"充值时间", "充值金额", "充值方式", "用户", "所属分组", "应得分红", "站长获利"})
	for _, order := range orders {
		if order == nil || order.PaidAt == nil {
			continue
		}
		info := promoInfo[order.UserID]
		amount := order.PayAmount
		dividend, profit := exportDividendAndProfit(amount, info.Code)
		user := order.UserEmail
		if strings.TrimSpace(order.UserName) != "" {
			user = order.UserName + " (" + order.UserEmail + ")"
		}
		group := info.Group
		if group == "" {
			group = "无优惠码"
		}
		_ = w.Write([]string{
			order.PaidAt.In(loc).Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", amount),
			paymentTypeLabel(order.PaymentType),
			user,
			group,
			fmt.Sprintf("%.2f", dividend),
			fmt.Sprintf("%.2f", profit),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	filename := fmt.Sprintf("payment-orders-%s-to-%s.csv", start.In(loc).Format("20060102"), end.Add(-time.Nanosecond).In(loc).Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(builder.String()))
}

// GetOrderDetail returns detailed information about a single order.
// GET /api/v1/admin/payment/orders/:id
func (h *PaymentHandler) GetOrderDetail(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	order, err := h.paymentService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	auditLogs, _ := h.paymentService.GetOrderAuditLogs(c.Request.Context(), orderID)
	promoInfo, _ := h.paymentService.GetRegistrationPromoInfo(c.Request.Context(), []int64{order.UserID})
	response.Success(c, gin.H{"order": adminPaymentOrderResponse(order, promoInfo[order.UserID]), "auditLogs": auditLogs})
}

func adminPaymentOrderResponses(orders []*dbent.PaymentOrder, infos map[int64]service.RegistrationPromoInfo) []map[string]any {
	out := make([]map[string]any, 0, len(orders))
	for _, order := range orders {
		if order != nil {
			out = append(out, adminPaymentOrderResponse(order, infos[order.UserID]))
		}
	}
	return out
}

func adminPaymentOrderResponse(order *dbent.PaymentOrder, info service.RegistrationPromoInfo) map[string]any {
	cloned := sanitizeAdminPaymentOrderForResponse(order)
	b, _ := json.Marshal(cloned)
	result := make(map[string]any)
	_ = json.Unmarshal(b, &result)
	if info.Code != "" {
		result["registration_promo_code"] = info.Code
		result["registration_promo_group"] = info.Group
	}
	return result
}

func parseExportDateRange(startRaw, endRaw string, loc *time.Location) (time.Time, time.Time, error) {
	if loc == nil {
		loc = time.Local
	}
	parse := func(raw string, endOfDay bool) (time.Time, error) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return time.Time{}, fmt.Errorf("start_date and end_date are required")
		}
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			return t, nil
		}
		if t, err := time.ParseInLocation("2006-01-02", raw, loc); err == nil {
			if endOfDay {
				return t.AddDate(0, 0, 1), nil
			}
			return t, nil
		}
		return time.Time{}, fmt.Errorf("invalid date: %s", raw)
	}
	start, err := parse(startRaw, false)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parse(endRaw, true)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must be later than start_date")
	}
	return start, end, nil
}

func exportLocation(raw string) *time.Location {
	if raw != "" {
		if loc, err := time.LoadLocation(raw); err == nil {
			return loc
		}
	}
	return time.Local
}

func paymentTypeLabel(value string) string {
	switch {
	case strings.HasPrefix(value, "wxpay"):
		return "微信支付"
	case strings.HasPrefix(value, "alipay"):
		return "支付宝"
	case strings.HasPrefix(value, "stripe"):
		return "Stripe"
	case strings.HasPrefix(value, "airwallex"):
		return "Airwallex"
	default:
		return value
	}
}

func exportDividendAndProfit(amount float64, promoCode string) (dividend, profit float64) {
	dividend = amount * 0.96
	code := strings.ToUpper(strings.TrimSpace(promoCode))
	if code == "" || code == "CHENLUREC" {
		return dividend, dividend
	}
	dividend *= 0.85
	return dividend, amount - dividend
}

// CancelOrder cancels a pending order (admin).
// POST /api/v1/admin/payment/orders/:id/cancel
func (h *PaymentHandler) CancelOrder(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	msg, err := h.paymentService.AdminCancelOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": msg})
}

// RetryFulfillment retries fulfillment for a paid order.
// POST /api/v1/admin/payment/orders/:id/retry
func (h *PaymentHandler) RetryFulfillment(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.paymentService.RetryFulfillment(c.Request.Context(), orderID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "fulfillment retried"})
}

func sanitizeAdminPaymentOrdersForResponse(orders []*dbent.PaymentOrder) []*dbent.PaymentOrder {
	if len(orders) == 0 {
		return orders
	}
	out := make([]*dbent.PaymentOrder, 0, len(orders))
	for _, order := range orders {
		out = append(out, sanitizeAdminPaymentOrderForResponse(order))
	}
	return out
}

func sanitizeAdminPaymentOrderForResponse(order *dbent.PaymentOrder) *dbent.PaymentOrder {
	if order == nil {
		return nil
	}
	cloned := *order
	cloned.ProviderSnapshot = nil
	return &cloned
}

// AdminProcessRefundRequest is the request body for admin refund processing.
type AdminProcessRefundRequest struct {
	Amount        float64 `json:"amount"`
	Reason        string  `json:"reason"`
	Force         bool    `json:"force"`
	DeductBalance bool    `json:"deduct_balance"`
}

// ProcessRefund processes a refund for an order (admin).
// POST /api/v1/admin/payment/orders/:id/refund
func (h *PaymentHandler) ProcessRefund(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req AdminProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	plan, earlyResult, err := h.paymentService.PrepareRefund(c.Request.Context(), orderID, req.Amount, req.Reason, req.Force, req.DeductBalance)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if earlyResult != nil {
		response.Success(c, earlyResult)
		return
	}

	result, err := h.paymentService.ExecuteRefund(c.Request.Context(), plan)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// --- Subscription Plans ---

// ListPlans returns all subscription plans.
// GET /api/v1/admin/payment/plans
func (h *PaymentHandler) ListPlans(c *gin.Context) {
	plans, err := h.configService.ListPlans(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, plans)
}

// CreatePlan creates a new subscription plan.
// POST /api/v1/admin/payment/plans
func (h *PaymentHandler) CreatePlan(c *gin.Context) {
	var req service.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	plan, err := h.configService.CreatePlan(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, plan)
}

// UpdatePlan updates an existing subscription plan.
// PUT /api/v1/admin/payment/plans/:id
func (h *PaymentHandler) UpdatePlan(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	plan, err := h.configService.UpdatePlan(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, plan)
}

// DeletePlan deletes a subscription plan.
// DELETE /api/v1/admin/payment/plans/:id
func (h *PaymentHandler) DeletePlan(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.configService.DeletePlan(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "deleted"})
}

// --- Provider Instances ---

// ListProviders returns all payment provider instances.
// GET /api/v1/admin/payment/providers
func (h *PaymentHandler) ListProviders(c *gin.Context) {
	providers, err := h.configService.ListProviderInstancesWithConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, providers)
}

// CreateProvider creates a new payment provider instance.
// POST /api/v1/admin/payment/providers
func (h *PaymentHandler) CreateProvider(c *gin.Context) {
	var req service.CreateProviderInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	inst, err := h.configService.CreateProviderInstance(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	h.paymentService.RefreshProviders(c.Request.Context())
	response.Created(c, inst)
}

// UpdateProvider updates an existing payment provider instance.
// PUT /api/v1/admin/payment/providers/:id
func (h *PaymentHandler) UpdateProvider(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdateProviderInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	inst, err := h.configService.UpdateProviderInstance(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	h.paymentService.RefreshProviders(c.Request.Context())
	response.Success(c, inst)
}

// DeleteProvider deletes a payment provider instance.
// DELETE /api/v1/admin/payment/providers/:id
func (h *PaymentHandler) DeleteProvider(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.configService.DeleteProviderInstance(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	h.paymentService.RefreshProviders(c.Request.Context())
	response.Success(c, gin.H{"message": "deleted"})
}

// parseIDParam parses an int64 path parameter.
// Returns the parsed ID and true on success; on failure it writes a BadRequest response and returns false.
func parseIDParam(c *gin.Context, paramName string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(paramName), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid "+paramName)
		return 0, false
	}
	return id, true
}

// --- Config ---

// GetConfig returns the payment configuration (admin view).
// GET /api/v1/admin/payment/config
func (h *PaymentHandler) GetConfig(c *gin.Context) {
	cfg, err := h.configService.GetPaymentConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// UpdateConfig updates the payment configuration.
// PUT /api/v1/admin/payment/config
func (h *PaymentHandler) UpdateConfig(c *gin.Context) {
	var req service.UpdatePaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.configService.UpdatePaymentConfig(c.Request.Context(), req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "updated"})
}
