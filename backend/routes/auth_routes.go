package routes

import (
	"final_project/controllers"

	"github.com/gofiber/fiber/v2"
)

func authRoutes(api fiber.Router) {
	api.Post("/auth/register", controllers.Register)
	api.Post("/auth/login", controllers.Login)
	api.Post("/auth/refresh", controllers.RefreshToken)
	api.Post("/auth/logout", controllers.Logout)
}
