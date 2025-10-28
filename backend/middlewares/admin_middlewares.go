package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"strings"
)

func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "forbidden: role not found"})
		}
		// role bisa berupa string atau jwt claim type; konversi aman ke string
		roleStr := ""
		switch v := role.(type) {
		case string:
			roleStr = v
		case []byte:
			roleStr = string(v)
		default:
			// jwt.MapClaims kadang menyimpan sebagai interface{} string
			roleStr = ""
		}
		if strings.ToLower(roleStr) != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "forbidden: admin only"})
		}
		return c.Next()
	}
}