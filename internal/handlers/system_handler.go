package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
	"github.com/google/uuid"
)

type SystemHandler struct {
	svc *services.SystemService
}

func NewSystemHandler(svc *services.SystemService) *SystemHandler {
	return &SystemHandler{svc: svc}
}

func (h *SystemHandler) SetupRoutes(router fiber.Router) {
	group := router.Group("/system")
	group.Use(middlewares.JWTMiddleware())
	group.Use(middlewares.TenantMiddleware())

	// Notifications
	group.Get("/notifications", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.GetNotifications)
	group.Put("/notifications/:id/read", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.MarkRead)
	group.Put("/notifications/read-all", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.MarkAllRead)
	group.Post("/notifications/call-staff", h.CallStaff)

	// Action Logs
	group.Get("/action-logs", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager), h.GetActionLogs)
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
