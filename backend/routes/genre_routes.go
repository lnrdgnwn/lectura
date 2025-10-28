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
	api.Post("/genres", controllers.CreateGenre, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	api.Put("/genres/:id", controllers.UpdateGenre, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
	api.Delete("/genres/:id", controllers.DeleteGenre, middlewares.AuthMiddleware(), middlewares.AdminMiddleware())
}
