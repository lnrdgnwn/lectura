package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func genreRoutes(api fiber.Router) {
	api.Get("/genres", controllers.ListGenres)
	api.Get("/genres/home", controllers.GetHomeGenres)
	api.Get("/genres/:id", controllers.GetGenre)

	api.Post("/genres", middlewares.AuthMiddleware(), middlewares.AdminMiddleware(), controllers.CreateGenre)
	api.Put("/genres/:id", middlewares.AuthMiddleware(), middlewares.AdminMiddleware(), controllers.UpdateGenre)
	api.Delete("/genres/:id", middlewares.AuthMiddleware(), middlewares.AdminMiddleware(), controllers.DeleteGenre)
	api.Put("/admin/genres/home", middlewares.AuthMiddleware(), middlewares.AdminMiddleware(), controllers.GenreShowByAdmin)
}
