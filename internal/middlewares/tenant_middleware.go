package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
)

// TenantMiddleware isolates sandbox contexts. It intercepts requests ensuring standard roles cannot bypass tenant_id.
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, okRole := c.Locals("role").(string)
		if !okRole {
			return c.Status(401).JSON(errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Missing role claim", nil))
		}

		systemRole, _ := c.Locals("system_role").(string)

		// Superadmin and Admin operate globally. Tenant filtering is optional or bypassed.
		if systemRole == string(domain.SystemRoleSuperuser) || systemRole == string(domain.SystemRoleAdmin) {
			return c.Next()
		}

		tenantID, ok := c.Locals("tenant_id").(string)
		if !ok || tenantID == "" {
			logger.Log.Error("CRITICAL: Tenant isolation breach attempted. Missing tenant_id in JWT.")
			return c.Status(403).JSON(errors.NewForbidden(errors.ErrCodeTenantIsolationBreach, "Hành động vượt quyền. Không xác định được chuỗi cửa hàng.", nil))
		}

		// The tenant_id is already in locals from JWT. It will be passed down to DB scope layer.
		return c.Next()
	}
}
