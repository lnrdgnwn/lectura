package middlewares

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
			return strings.Trim(parts[1], `"`) // hapus quotes bila ada
		}
	}
	// 2. Cookie fallback (jika FE menyimpan token di cookie)
	if val := c.Cookies("access_token"); val != "" {
		return strings.Trim(val, `"`)
	}
	return ""
}

// AuthMiddleware memverifikasi access JWT dan menaruh informasi user di c.Locals
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "server misconfigured (missing JWT_SECRET)"})
		}

		tokenStr := strings.TrimSpace(extractToken(c))
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "missing access token"})
		}

		// Parse with explicit MapClaims so we can introspect validation errors
		token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
			// pastikan method signing HMAC (HS256, HS512, etc.)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil {
			// lebih detail saat debugging: cek jenis error
			if ve, ok := err.(*jwt.ValidationError); ok {
				// You can log ve.Errors to see flags
				if ve.Errors&jwt.ValidationErrorExpired != 0 {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "token expired"})
				}
				if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token signature"})
				}
				// fallback: generic invalid token
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token", "error": ve.Error()})
			}
			// other parse error
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token", "error": err.Error()})
		}

		if token == nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid or expired token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token claims"})
		}

		// (opsional) Validasi exp secara eksplisit (kadang berguna)
		if expVal, ok := claims["exp"]; ok {
			switch exp := expVal.(type) {
			case float64:
				if time.Unix(int64(exp), 0).Before(time.Now()) {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "token expired"})
				}
			}
		}

		// Ambil sub (user id) -> normalisasi ke uint jika memungkinkan
		var rawSub interface{}
		if v, exists := claims["sub"]; exists {
			rawSub = v
		} else if v, exists := claims["user_id"]; exists {
			rawSub = v
		}

		if rawSub == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "token missing subject"})
		}

		// normalize: try float64 -> uint, string numeric -> uint, else keep raw
		switch t := rawSub.(type) {
		case float64:
			c.Locals("user_id", uint(t))
		case int:
			c.Locals("user_id", uint(t))
		case int64:
			c.Locals("user_id", uint(t))
		case uint:
			c.Locals("user_id", t)
		case string:
			// coba parse numeric string
			if n, err := strconv.Atoi(t); err == nil {
				c.Locals("user_id", uint(n))
			} else {
				// simpan string jika tidak numeric (controller harus handle)
				c.Locals("user_id", t)
			}
		default:
			c.Locals("user_id", t)
		}

		// role
		if r, exists := claims["role"]; exists {
			if rs, ok := r.(string); ok {
				c.Locals("role", rs)
			} else {
				c.Locals("role", r)
			}
		} else {
			c.Locals("role", nil)
		}

		return c.Next()
	}
}
