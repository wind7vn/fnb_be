package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
)

// BaseResponse represents the standard success response payload structure.
type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Success returns a standardized 200 OK HTTP response with payload data.
func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(BaseResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage returns a standardized 200 OK HTTP response with context message and optional data.
func SuccessWithMessage(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created returns a standardized 201 Created HTTP response with payload.
func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(BaseResponse{
		Success: true,
		Data:    data,
	})
}

// Error returns a standardized error response based on AppError.
func Error(c *fiber.Ctx, err *errors.AppError) error {
	return c.Status(err.StatusCode).JSON(err)
}
