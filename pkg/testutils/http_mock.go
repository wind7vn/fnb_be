package testutils

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
)

// MockJSONRequest generates a mock HTTP request that seamlessly feeds into testing Fiber handlers
func MockJSONRequest(method string, route string, body []byte) *http.Request {
	var req *http.Request

	if body != nil {
		req = httptest.NewRequest(method, route, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, route, nil)
	}

	req.Header.Set("Content-Type", "application/json")
	return req
}

// ReadResponse retrieves the output body byte slice returned by the Fiber test engine
func ReadResponse(resp *http.Response) []byte {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body
}

// SetupMockApp builds a fresh Fiber instance for testing controllers safely
func SetupMockApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		},
	})
}
