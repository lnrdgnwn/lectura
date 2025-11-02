package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"strings"
)

func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden: Your role not found"})
		}
		roleStr := ""
		switch v := role.(type) {
		case string:
			roleStr = v
		case []byte:
			roleStr = string(v)
		default:
			roleStr = ""
		}
		if strings.ToLower(roleStr) != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "forbidden: admin only"})
		}
		return c.Next()
	}
}