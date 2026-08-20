package handlers

import "github.com/gofiber/fiber/v2"

// Response đại diện cho cấu trúc phản hồi API chuẩn hoá.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success gửi phản hồi thành công.
func success(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Data:    data,
	})
}

// ErrResponse gửi phản hồi lỗi.
func errResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Error:   message,
	})
}

func unauthorized(c *fiber.Ctx) error {
	return errResponse(c, fiber.StatusUnauthorized, "unauthorized")
}

func internalError(c *fiber.Ctx) error {
	return errResponse(c, fiber.StatusInternalServerError, "internal server error")
}
