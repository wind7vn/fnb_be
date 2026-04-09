package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
)

type ProductHandler struct {
	svc *services.ProductService
}

func NewProductHandler(svc *services.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) SetupRoutes(router fiber.Router) {
	prodGroup := router.Group("/products")
	
	prodGroup.Use(middlewares.JWTMiddleware())
	prodGroup.Use(middlewares.TenantMiddleware())

	// Read is public to all staff inside the tenant
	prodGroup.Get("/", h.GetAll)

	// Write is restricted to Managers and Owners
	writeGroup := prodGroup.Group("/", middlewares.RolesAllowed(domain.RoleOwner, domain.RoleManager))
	writeGroup.Post("/", h.Create)
	writeGroup.Put("/:id", h.Update)
	writeGroup.Delete("/:id", h.Delete)
}

func (h *ProductHandler) Create(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	var req services.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid format", err))
	}

	prod, appErr := h.svc.Create(tenantID, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.Created(c, prod)
}

func (h *ProductHandler) GetAll(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	category := c.Query("category")

	var catPtr *string
	if category != "" {
		catPtr = &category
	}

	products, appErr := h.svc.GetAll(tenantID, catPtr)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.Success(c, products)
}

func (h *ProductHandler) Update(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	productID := c.Params("id")

	var req services.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid format", err))
	}

	prod, appErr := h.svc.Update(tenantID, productID, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}
	return response.Success(c, prod)
}

func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	productID := c.Params("id")

	if appErr := h.svc.Delete(tenantID, productID); appErr != nil {
		return response.Error(c, appErr)
	}
	return response.SuccessWithMessage(c, "Deleted successfully", nil)
}
