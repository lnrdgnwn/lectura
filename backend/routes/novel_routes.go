package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func novelRoutes(api fiber.Router) {
	// public
	api.Get("/novels", controllers.GetNovels)
	api.Get("/novels/search", controllers.SearchNovels)
	api.Get("/novels/:id", controllers.GetNovel)
	api.Get("/novels/:novel_id/chapters", controllers.ListChapters)

	// protected (create / modify / delete) - require auth + admin
	api.Post("/novels", controllers.PostNovel, middlewares.AdminMiddleware(), middlewares.AuthMiddleware())
	api.Put("/novels/:id", controllers.UpdateNovel, middlewares.AdminMiddleware(), middlewares.AuthMiddleware())
	api.Delete("/novels/:id", controllers.DeleteNovel, middlewares.AdminMiddleware(), middlewares.AuthMiddleware())

	// chapters - protected for creation
	api.Post("/chapters", controllers.AddChapter, middlewares.AdminMiddleware(), middlewares.AuthMiddleware())
}
