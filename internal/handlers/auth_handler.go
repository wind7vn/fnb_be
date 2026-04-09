package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) SetupRoutes(router fiber.Router) {
	authGroup := router.Group("/auth")

	// Public
	authGroup.Post("/login", h.Login)

	// Protected
	authGroup.Use(middlewares.JWTMiddleware())
	authGroup.Get("/me", h.GetMe)
	authGroup.Post("/devices", h.RegisterDevice)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req services.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid parsing format", err))
	}

	resp, appErr := h.authService.Login(req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, resp)
}

// POST /auth/devices
// { "device_id": "XY", "fcm_token": "xxx", "platform": "ios" }
func (h *AuthHandler) RegisterDevice(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req struct {
		DeviceID string `json:"device_id"`
		FCMToken string `json:"fcm_token"`
		Platform string `json:"platform"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Invalid parsing format", err))
	}

	if appErr := h.authService.RegisterDevice(userID, req.DeviceID, req.FCMToken, req.Platform); appErr != nil {
		return response.Error(c, appErr)
	}

	return response.SuccessWithMessage(c, "Device registered successfully", nil)
}

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, appErr := h.authService.GetMe(userID)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, fiber.Map{
		"id":           user.ID,
		"full_name":    user.FullName,
		"phone_number": user.PhoneNumber,
		"role":         user.Role,
		"avatar_url":   user.AvatarURL,
	})
}
