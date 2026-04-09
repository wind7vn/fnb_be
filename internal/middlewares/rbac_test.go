package middlewares_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/wind7vn/fnb_be/internal/middlewares"
)

func TestRolesAllowed_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("role", "Owner")
		return c.Next()
	}, middlewares.RolesAllowed("Owner", "Manager"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestRolesAllowed_Forbidden(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("role", "Staff")
		return c.Next()
	}, middlewares.RolesAllowed("Owner", "Manager"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 403, resp.StatusCode)
}
