package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
)

type TenantHandler struct {
	tenantService *services.TenantService
	aiService     *services.AIService
}

func NewTenantHandler(tenantService *services.TenantService, aiService *services.AIService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService, aiService: aiService}
}

func (h *TenantHandler) SetupRoutes(router fiber.Router) {
	sysGroup := router.Group("/system")
	
	sysGroup.Use(middlewares.JWTMiddleware())
	// Only Superadmin or Admin can manage system-wide tenants
	sysGroup.Use(middlewares.RolesAllowed(
		domain.RoleSuperadmin, string(domain.SystemRoleSuperuser),
		domain.RoleAdmin, string(domain.SystemRoleAdmin),
	))

	sysGroup.Post("/tenants", h.CreateTenant)
	sysGroup.Post("/admins", h.CreateAdmin)
	sysGroup.Get("/tenants", h.GetTenants)

	// Owner/Manager APIs scoped by Tenant sandbox
	tenantGroup := router.Group("/tenant")
	tenantGroup.Use(middlewares.JWTMiddleware())
	tenantGroup.Use(middlewares.TenantMiddleware())

	tenantGroup.Get("/staff", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager), h.GetStaff)
	tenantGroup.Post("/staff", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager), h.CreateStaff)
	tenantGroup.Put("/settings", middlewares.RolesAllowed(domain.RoleOwner), h.UpdateSettings)
	tenantGroup.Post("/ai/scan-menu", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager), h.ScanMenuByAI)
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

	return response.Success(c, tenant)
}

func (h *TenantHandler) CreateAdmin(c *fiber.Ctx) error {
	var req services.CreateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Dữ liệu không hợp lệ", err))
	}

	adminUser, appErr := h.tenantService.CreateSystemAdmin(req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, adminUser)
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
		"role":         req.Role,
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

func (h *TenantHandler) ScanMenuByAI(c *fiber.Ctx) error {
	fmt.Println("----> DEBUG: API Hitted!")
	file, err := c.FormFile("image")
	if err != nil {
		fmt.Println("----> DEBUG: Error FormFile", err)
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Mising image file", err))
	}

	// Read image bytes
	fileContent, err := file.Open()
	if err != nil {
		fmt.Println("----> DEBUG: Error file open", err)
		return response.Error(c, errors.NewInternalServer(err))
	}
	defer fileContent.Close()

	buf := make([]byte, file.Size)
	if _, err := fileContent.Read(buf); err != nil {
		fmt.Println("----> DEBUG: Error Read file size: ", file.Size, err)
		return response.Error(c, errors.NewInternalServer(err))
	}

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	fmt.Printf("----> DEBUG: Image Size: %d bytes, MIME: %s\n", len(buf), mimeType)

	items, appErr := h.aiService.ExtractMenuFromImage(buf, mimeType)
	if appErr != nil {
		fmt.Println("----> DEBUG: Ai Err:", appErr.Message)
		return response.Error(c, appErr)
	}

	fmt.Println("----> DEBUG: Success!")
	return response.Success(c, items)
}
