// user_routes.go
package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func userRoutes(api fiber.Router) {
	// profile (requires login)
	api.Get("/users/me", middlewares.AuthMiddleware(), controllers.GetMe)
	api.Put("/users/me", middlewares.AuthMiddleware(), controllers.UpdateProfile)
	api.Put("/users/me/password", middlewares.AuthMiddleware(), controllers.ChangePassword)

	// admin-only user management
	api.Get("/admin/users", middlewares.AuthMiddleware(), middlewares.AdminMiddleware(),controllers.ListUsers)
	api.Delete("/admin/users/:id",  middlewares.AuthMiddleware(), middlewares.AdminMiddleware(),controllers.DeleteUser)
}
