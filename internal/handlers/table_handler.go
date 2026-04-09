package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
)

type TableHandler struct {
	svc *services.TableService
}

func NewTableHandler(svc *services.TableService) *TableHandler {
	return &TableHandler{svc: svc}
}

func (h *TableHandler) SetupRoutes(router fiber.Router) {
	tableGroup := router.Group("/tables")
	
	tableGroup.Use(middlewares.JWTMiddleware())
	tableGroup.Use(middlewares.TenantMiddleware())

	// Staff+ can Read and Modify Status (via workflow)
	tableGroup.Get("/", h.GetAll)
	tableGroup.Put("/:id/status", h.UpdateStatus)

	// Only Owner/Manager can Create tables structurally
	tableGroup.Post("/", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager), h.Create)
}

func (h *TableHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	var req services.TableRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid format", err))
	}

	table, appErr := h.svc.Create(tenantID, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.Created(c, table)
}

func (h *TableHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	tables, appErr := h.svc.GetAll(tenantID)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.Success(c, tables)
}

func (h *TableHandler) UpdateStatus(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	tableID := c.Params("id")

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid format", err))
	}

	table, appErr := h.svc.UpdateStatus(tenantID, tableID, req.Status)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.Success(c, table)
}
