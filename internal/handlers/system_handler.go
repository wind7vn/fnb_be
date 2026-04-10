package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
	"github.com/wind7vn/fnb_be/pkg/config"
)

type SystemHandler struct {
	svc *services.SystemService
}

func NewSystemHandler(svc *services.SystemService) *SystemHandler {
	return &SystemHandler{svc: svc}
}

func (h *SystemHandler) SetupRoutes(router fiber.Router) {
	group := router.Group("/system")

	// Public Config Route
	group.Get("/configs", h.GetClientConfigs)

	group.Use(middlewares.JWTMiddleware())
	group.Use(middlewares.TenantMiddleware())

	// Notifications
	group.Get("/notifications", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.GetNotifications)
	group.Put("/notifications/:id/read", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.MarkRead)
	group.Put("/notifications/read-all", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.MarkAllRead)
	group.Post("/notifications/call-staff", h.CallStaff)

	// Action Logs
	group.Get("/action-logs", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager), h.GetActionLogs)

	// Banks Configuration
	group.Get("/banks", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.GetBanks)
}

func (h *SystemHandler) CallStaff(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	var req struct {
		TableID   string `json:"table_id"`
		TableName string `json:"table_name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid payload", err))
	}

	title := req.TableName + " gọi phục vụ"
	message := req.TableName + " vừa nhấn chuông yêu cầu hỗ trợ"

	data := map[string]interface{}{
		"table_id": req.TableID,
		"action":   "CALL_STAFF",
	}

	h.svc.CreateNotification(tenantID, uuid.Nil.String(), title, message, "CALL_STAFF", data)

	return response.Success(c, "Đã gửi thông báo cho nhân viên")
}

func (h *SystemHandler) GetNotifications(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	notis, appErr := h.svc.GetUnreadNotifications(tenantID, limit)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, notis)
}

func (h *SystemHandler) MarkRead(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	id := c.Params("id")

	appErr := h.svc.MarkRead(tenantID, id)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, nil)
}

func (h *SystemHandler) MarkAllRead(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	appErr := h.svc.MarkAllRead(tenantID)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, nil)
}

func (h *SystemHandler) GetActionLogs(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logs, appErr := h.svc.GetLogs(tenantID, limit, offset)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, logs)
}

func (h *SystemHandler) GetClientConfigs(c *fiber.Ctx) error {
	domainUrl := config.AppConfig.AppDomain
	if domainUrl == "" {
		domainUrl = "http://localhost:8080"
	}

	wsUrl := strings.Replace(domainUrl, "https://", "wss://", 1)
	wsUrl = strings.Replace(wsUrl, "http://", "ws://", 1)

	configs := map[string]interface{}{
		"base_url":               domainUrl,
		"api_base_url":           domainUrl + "/api/v1",
		"socket_base_url":        wsUrl + "/api/v1/ws",
		"term_url":               domainUrl + "/support/terms-of-service",
		"privacy_url":            domainUrl + "/support/privacy-policy",
		"phone_center":           "19001234",
		"max_wait_pong":          2000,
		"max_delay_ping":         10000,
		"check_net_domain":       "https://www.google.com",
		"max_delay_check_net":    5000,
		"max_recent_destination": 7,
	}

	return c.JSON(configs)
}

func (h *SystemHandler) GetBanks(c *fiber.Ctx) error {
	banks, err := h.svc.GetSupportedBanks()
	if err != nil {
		return response.Error(c, err)
	}
	return response.Success(c, banks)
}

