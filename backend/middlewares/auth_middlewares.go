package middlewares

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// Ambil token dari header Authorization atau cookie "access_token"
func extractToken(c *fiber.Ctx) string {
	// 1. Authorization header
	auth := c.Get("Authorization")
	if auth != "" {
		parts := strings.Fields(auth) // ["Bearer", "token"]
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	// 2. Cookie fallback (jika FE menyimpan token di cookie)
	if val := c.Cookies("access_token"); val != "" {
		return val
	}

	return ""
}

// AuthMiddleware memverifikasi access JWT dan menaruh informasi user di c.Locals
// - Mengembalikan 401 bila token tidak ada/invalid/expired
// - Menaruh user_id (uint) pada c.Locals("user_id") dan role (string) pada c.Locals("role")
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			// konfigurasi JWT belum ada
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "server misconfigured (missing JWT_SECRET)"})
		}

		tokenStr := extractToken(c)
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "missing access token"})
		}

		// Parse token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// pastikan alg adalah HMAC
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(secret), nil
		})
		if err != nil || token == nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid or expired token"})
		}

		// Ambil claims (MapClaims)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token claims"})
		}

		// Ambil sub (user id) dan role dari claims; variasi: sub bisa numeric atau string
		var userID interface{}
		if v, exists := claims["sub"]; exists {
			userID = v
		} else if v, exists := claims["user_id"]; exists { // fallback
			userID = v
		}

		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "token missing subject"})
		}

		// simpan secara langsung (tipe bisa float64 atau string)
		c.Locals("user_id", userID)

		// ambil role jika ada
		if r, exists := claims["role"]; exists {
			c.Locals("role", r)
		} else {
			c.Locals("role", nil)
		}

		// lanjut ke handler berikutnya
		return c.Next()
	}
}

