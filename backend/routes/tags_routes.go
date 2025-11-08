package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func tagsRoutes(api fiber.Router) {
	api.Get("/tags", controllers.ListTags)
	api.Get("/tags/:id", controllers.GetTag)

	api.Post("/tags", middlewares.AuthMiddleware(),middlewares.AdminMiddleware(), controllers.CreateTag)
	api.Put("/tags/:id", middlewares.AuthMiddleware(),middlewares.AdminMiddleware(), controllers.UpdateTag)
	api.Delete("/tags/:id", middlewares.AuthMiddleware(),middlewares.AdminMiddleware(), controllers.DeleteTag)
}
