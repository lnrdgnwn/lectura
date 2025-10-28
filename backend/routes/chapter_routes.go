package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func chapterRoutes(api fiber.Router) {
	// public
	api.Get("/novels/:novel_id/chapters", controllers.ListChapters)
	api.Get("/chapters/:id", controllers.GetChapter)

	// protected (create / modify / delete) - require auth + admin
	api.Post("/chapters", controllers.AddChapter, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	api.Put("/chapters/:id", controllers.UpdateChapter, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	api.Delete("/chapters/:id", controllers.DeleteChapter, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
}
