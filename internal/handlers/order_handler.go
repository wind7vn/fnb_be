package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
)

type OrderHandler struct {
	svc *services.OrderService
}

func NewOrderHandler(svc *services.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) SetupRoutes(router fiber.Router) {
	orderGroup := router.Group("/orders")
	orderGroup.Use(middlewares.JWTMiddleware())
	orderGroup.Use(middlewares.TenantMiddleware())

	// Create/Open table session (generates Order)
	// Open to Staff/Owner and Guests possessing exact Table QR token
	orderGroup.Post("/", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff, "Guest"), h.OpenSession)
	// Allow Guest as well to load active order (for their UI to work)
	orderGroup.Get("/tables/:tableId/active", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff, "Guest"), h.GetActiveOrderByTable)

	// Open to Staff/Owner and Guests possessing exact Table QR token
	orderGroup.Put("/:id/items", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff, "Guest"), h.AddItems) 

	// Checkout is Staff only
	orderGroup.Put("/:id/checkout", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.Checkout)

	// Generate QR token. Staff side mechanism
	orderGroup.Post("/guest", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.GenerateGuestQR)

	// KDS specific updating Endpoint
	orderGroup.Put("/items/:itemId/status", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager, domain.RoleStaff), h.UpdateItemStatus)
}

func (h *OrderHandler) GetActiveOrderByTable(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	tableID := c.Params("tableId")

	order, appErr := h.svc.GetActiveOrderByTable(tenantID, tableID)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	if role := c.Locals("role"); role == "Guest" {
		clientSession := c.Get("X-Table-Session-Token")
		if order.SessionID != "" && clientSession != order.SessionID {
			return response.Error(c, errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Bàn đang có khách. Vui lòng liên hệ nhân viên.", nil))
		}
	}

	return response.Success(c, order)
}

func (h *OrderHandler) OpenSession(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	tableID := c.Params("tableId")

	var req services.OpenSessionRequest
	// parse body if any
	_ = c.BodyParser(&req)

	// Fallback to URL parameter
	if req.TableID == "" {
		req.TableID = tableID
	}

	// For guest, inject session_id from header if not in body
	if role := c.Locals("role"); role == "Guest" {
		if req.SessionID == "" {
			req.SessionID = c.Get("X-Table-Session-Token")
		}
		if req.SessionID == "" {
			req.SessionID = uuid.New().String()
		}
	}

	order, appErr := h.svc.OpenSession(tenantID, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Created(c, order)
}

func (h *OrderHandler) AddItems(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	orderID := c.Params("id")

	var req struct {
		SessionID string `json:"session_id"`
		Items []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		} `json:"items"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid format", err))
	}

	sessionID := req.SessionID
	if role := c.Locals("role"); role == "Guest" {
		if sessionID == "" {
			sessionID = c.Get("X-Table-Session-Token")
		}
		
		tokenTableID, ok := c.Locals("table_id").(string)
		if !ok || tokenTableID == "" {
			return response.Error(c, errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Guest token invalid", nil))
		}
		
		// Verify order belongs to guest's table
		order, err := h.svc.GetActiveOrderByTable(tenantID, tokenTableID)
		if err != nil || order.ID.String() != orderID {
			return response.Error(c, errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Cannot access this order", nil))
		}

		if order.SessionID != "" && sessionID != order.SessionID {
			return response.Error(c, errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Bàn đang được thanh toán.", nil))
		}
	}

	order, appErr := h.svc.AddItems(tenantID, orderID, sessionID, req.Items)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, order)
}

func (h *OrderHandler) GenerateGuestQR(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	
	var req struct {
		TableID string `json:"table_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid format", err))
	}

	token, appErr := h.svc.GenerateGuestToken(tenantID, req.TableID)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, map[string]string{
		"guest_token": token,
	})
}

func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	orderID := c.Params("id")

	order, appErr := h.svc.Checkout(tenantID, orderID)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, order)
}

func (h *OrderHandler) UpdateItemStatus(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	itemID := c.Params("itemId")

	var req struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid body format", err))
	}

	appErr := h.svc.UpdateItemStatus(tenantID, itemID, req.Status)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, nil)
}
