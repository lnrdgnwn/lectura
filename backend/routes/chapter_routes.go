package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func chapterRoutes(api fiber.Router) {
	api.Get("/chapters", controllers.GetAllChapters)
	api.Get("/chapters/:id", controllers.GetChapter)

	api.Post("/chapters", middlewares.AuthMiddleware(), controllers.AddChapter)
	api.Put("/chapters/:id", middlewares.AuthMiddleware(), controllers.UpdateChapter)
	api.Delete("/chapters/:id", middlewares.AuthMiddleware(), controllers.DeleteChapter)
}
