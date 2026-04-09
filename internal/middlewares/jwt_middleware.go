package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/config"
)

// UserClaims defines the payload inside the JWT Token
type UserClaims struct {
	UserID     string `json:"user_id"`
	TenantID   string `json:"tenant_id"` // Empty if Superadmin
	Role       string `json:"role"`
	SystemRole string `json:"system_role"`
	TableID    string `json:"table_id,omitempty"` // For Guest QR tokens
	jwt.RegisteredClaims
}

func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		tokenString := ""

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if tokenString == "" {
			tokenString = c.Query("token") // Support for WebSocket "?token="
		}

		if tokenString == "" {
			return c.Status(401).JSON(errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Vui lòng đăng nhập để tiếp tục", nil))
		}

		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Phiên đăng nhập không hợp lệ hoặc đã hết hạn", err))
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			return c.Status(401).JSON(errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Dữ liệu xác thực bị hỏng", nil))
		}

		// Inject Claims into Locals for subsequent middlewares
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("system_role", claims.SystemRole)
		c.Locals("tenant_id", claims.TenantID)
		if claims.TableID != "" {
			c.Locals("table_id", claims.TableID)
		}

		return c.Next()
	}
}
