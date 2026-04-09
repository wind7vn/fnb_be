package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
)

// RolesAllowed restricts the enclosed route group to a variadic list of accepted roles.
func RolesAllowed(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return c.Status(401).JSON(errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Nhận dạng quyền thất bại", nil))
		}

		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next() // Role is authorized
			}
		}

		// Deny
		return c.Status(403).JSON(errors.NewForbidden(errors.ErrCodeForbidden, "Bạn không có quyền truy cập tính năng này", nil))
	}
}
