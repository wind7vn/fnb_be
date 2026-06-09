package middlewares

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
)

// RolesAllowed restricts the enclosed route group to a variadic list of accepted roles.
// It checks both the tenant-scoped 'role' and the global 'system_role'.
func RolesAllowed(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, _ := c.Locals("role").(string)
		systemRole, _ := c.Locals("system_role").(string)

		logger.Log.Info(fmt.Sprintf("[RBAC] Checking access to path: %s. userRole: '%s', systemRole: '%s', allowedRoles: %v", c.Path(), userRole, systemRole, allowedRoles))

		if userRole == "" && systemRole == "" {
			return c.Status(401).JSON(errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Nhận dạng quyền thất bại", nil))
		}

		for _, role := range allowedRoles {
			if userRole == role || systemRole == role {
				return c.Next() // Role is authorized
			}
		}

		logger.Log.Warn(fmt.Sprintf("[RBAC] DENIED access to path: %s. userRole: '%s', systemRole: '%s', allowedRoles: %v", c.Path(), userRole, systemRole, allowedRoles))

		// Deny
		return c.Status(403).JSON(errors.NewForbidden(errors.ErrCodeForbidden, "Bạn không có quyền truy cập tính năng này", nil))
	}
}
