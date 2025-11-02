// middlewares/auth_middleware.go
package middlewares

import (
	"os"	

	"final_project/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func getRawAccessToken(c *fiber.Ctx) (string, error) {
	enc := c.Cookies("access_token")
	if enc == "" {
		return "", fiber.ErrUnauthorized
	}
	dec, err := utils.Decrypt(enc)
	if err != nil || dec == "" {
		return "", fiber.ErrUnauthorized
	}
	return dec, nil
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "server misconfigured: missing JWT_SECRET",
			})
		}

		raw, err := getRawAccessToken(c)
		if err != nil || raw == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
		}

		token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})
		if err != nil || token == nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token"})
		}

		var userID any
		if v, ok := claims["id"]; ok {
			userID = v
		} else if v, ok := claims["sub"]; ok {
			userID = v
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token (no id)"})
		}

		role, _ := claims["role"].(string)

		c.Locals("id", userID)    
		c.Locals("role", role)    

		return c.Next()
	}
}
