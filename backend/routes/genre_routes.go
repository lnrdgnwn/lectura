package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func genreRoutes(api fiber.Router) {
	// public
	api.Get("/genres", controllers.ListGenres)
	api.Get("/genres/:id", controllers.GetGenre)

	// protected: create/update/delete require auth + admin
	api.Post("/genres", middlewares.AuthMiddleware(), controllers.CreateGenre)
	api.Put("/genres/:id", middlewares.AuthMiddleware(), controllers.UpdateGenre)
	api.Delete("/genres/:id", middlewares.AuthMiddleware(), controllers.DeleteGenre)
}
