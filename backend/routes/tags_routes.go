package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func tagsRoutes(api fiber.Router) {
	// public
	api.Get("/tags", controllers.ListTags)
	api.Get("/tags/:id", controllers.GetTag)

	// admin-only tag management
	api.Post("/tags", middlewares.AuthMiddleware(), controllers.CreateTag)
	api.Put("/tags/:id", middlewares.AuthMiddleware(), controllers.UpdateTag)
	api.Delete("/tags/:id", middlewares.AuthMiddleware(), controllers.DeleteTag)
}
