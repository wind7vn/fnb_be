package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
)

type TenantHandler struct {
	tenantService *services.TenantService
}

func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

func (h *TenantHandler) SetupRoutes(router fiber.Router) {
	sysGroup := router.Group("/system")
	
	sysGroup.Use(middlewares.JWTMiddleware())
	// Only Superadmin or Admin can manage system-wide tenants
	sysGroup.Use(middlewares.RolesAllowed(domain.RoleSuperadmin, domain.RoleAdmin))

	sysGroup.Post("/tenants", h.CreateTenant)
	sysGroup.Get("/tenants", h.GetTenants)

	// Owner/Manager APIs scoped by Tenant sandbox
	tenantGroup := router.Group("/tenant")
	tenantGroup.Use(middlewares.JWTMiddleware())
	tenantGroup.Use(middlewares.TenantMiddleware())

	tenantGroup.Get("/staff", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager), h.GetStaff)
	tenantGroup.Post("/staff", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager), h.CreateStaff)
	tenantGroup.Put("/settings", middlewares.RolesAllowed(domain.RoleOwner), h.UpdateSettings)
}

func (h *TenantHandler) CreateTenant(c *fiber.Ctx) error {
	var req services.CreateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid request format", err))
	}

	tenant, appErr := h.tenantService.CreateTenant(req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Created(c, tenant)
}

func (h *TenantHandler) GetTenants(c *fiber.Ctx) error {
	tenants, appErr := h.tenantService.GetAllTenants()
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, tenants)
}

func (h *TenantHandler) CreateStaff(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	var req services.CreateStaffRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid request format", err))
	}

	staff, appErr := h.tenantService.CreateStaff(tenantID, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Created(c, map[string]string{
		"id":           staff.ID.String(),
		"full_name":    staff.FullName,
		"phone_number": staff.PhoneNumber,
		"role":         staff.Role,
	})
}

func (h *TenantHandler) GetStaff(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	staff, appErr := h.tenantService.GetStaff(tenantID)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, staff)
}

func (h *TenantHandler) UpdateSettings(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	var req services.UpdateSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid request format", err))
	}

	if appErr := h.tenantService.UpdateSettings(tenantID, req); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.SuccessWithMessage(c, "Settings updated successfully", nil)
}

func (h *TenantHandler) GetSettings(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	metadata, appErr := h.tenantService.GetSettings(tenantID)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	// Just return as metadata JSON object
	return response.Success(c, fiber.Map{"metadata": metadata})
}
