// user_routes.go
package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func userRoutes(api fiber.Router) {
	// profile (requires login)
	api.Get("/users/me", controllers.GetMe, middlewares.AuthMiddleware())
	api.Put("/users/me", controllers.UpdateProfile, middlewares.AuthMiddleware())

	// admin-only user management
	api.Get("/admin/users", controllers.ListUsers, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	api.Delete("/admin/users/:id", controllers.DeleteUser, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
}
