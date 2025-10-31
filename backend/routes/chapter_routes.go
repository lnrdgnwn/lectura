package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func chapterRoutes(api fiber.Router) {
	// public
	api.Get("/chapters", middlewares.AdminMiddleware(),controllers.GetAllChapters)
	api.Get("/chapters/:id", controllers.GetChapter)

	// protected (create / modify / delete) - require auth + admin
	api.Post("/chapters", middlewares.AuthMiddleware(), controllers.AddChapter)
	api.Put("/chapters/:id", middlewares.AuthMiddleware(), controllers.UpdateChapter)
	api.Delete("/chapters/:id", middlewares.AuthMiddleware(), controllers.DeleteChapter)
}
