// middlewares/admin_or_self.go
package middlewares

import (
	"os"
	"strconv"
	"strings"

	"final_project/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func authFromAccessCookie(c *fiber.Ctx) (uint, string, error) {
	enc := c.Cookies("access_token", "")
	if enc == "" {
		return 0, "", fiber.ErrUnauthorized
	}
	tokenStr, err := utils.Decrypt(enc)
	if err != nil || strings.TrimSpace(tokenStr) == "" {
		return 0, "", fiber.ErrUnauthorized
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return 0, "", fiber.ErrUnauthorized
	}

	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(secret), nil
	})
	if err != nil || tok == nil || !tok.Valid {
		return 0, "", fiber.ErrUnauthorized
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", fiber.ErrUnauthorized
	}

	var uid uint
	switch v := claims["id"].(type) {
	case float64:
		uid = uint(v)
	case string:
		if n, e := strconv.Atoi(v); e == nil {
			uid = uint(n)
		} else {
			return 0, "", fiber.ErrUnauthorized
		}
	default:
		return 0, "", fiber.ErrUnauthorized
	}

	role := ""
	if r, ok := claims["role"].(string); ok {
		role = strings.ToUpper(r)
	}
	return uid, role, nil
}

// AdminOrSelf: izinkan jika (param :id == userID) ATAU (role == ADMIN)
func AdminOrSelf(paramName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, role, err := authFromAccessCookie(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
		}

		// Ambil target ID dari route param (boleh "id" / "user_id" sesuai paramName)
		idStr := strings.TrimSpace(c.Params(paramName))
		var targetID int
		if idStr != "" {
			if n, e := strconv.Atoi(idStr); e == nil && n > 0 {
				targetID = n
			}
		}

		// 1) SELF: kalau id param valid & sama -> langsung lolos
		if targetID > 0 && uid == uint(targetID) {
			c.Locals("user_id", uid)
			c.Locals("role", role)
			return c.Next()
		}

		// 2) ADMIN: kalau admin -> langsung lolos (meski id tidak sama/atau param kosong)
		if role == "ADMIN" {
			c.Locals("user_id", uid)
			c.Locals("role", role)
			return c.Next()
		}

		// 3) Selain itu -> forbidden
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "forbidden"})
	}
}
