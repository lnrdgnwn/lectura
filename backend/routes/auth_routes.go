// auth_routes.go
package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func authRoutes(api fiber.Router) {
	// Auth endpoints (public)
	api.Post("/auth/register", controllers.Register)
	api.Post("/auth/login", controllers.Login)
	api.Post("/auth/refresh", controllers.Refresh)

	// Logout requires valid refresh token (we keep it protected by auth middleware if you store access token)
	// If you expect logout to be called with refresh_token only (no access token), you can keep this public.
	// Here we require authentication (access token) for logout.
	api.Post("/auth/logout", controllers.Logout, middlewares.AuthMiddleware())
}
