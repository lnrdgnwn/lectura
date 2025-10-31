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
	api.Post("/tags", controllers.CreateTag, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	api.Put("/tags/:id", controllers.UpdateTag, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	api.Delete("/tags/:id", controllers.DeleteTag, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())

	// manage tag associations on novels (require auth; you can change to Admin only if ingin)
	api.Post("/novels/:id/tags", controllers.AddTagsToNovel, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	api.Delete("/novels/:id/tags", controllers.RemoveTagsFromNovel, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
}
