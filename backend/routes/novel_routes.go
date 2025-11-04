package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func novelRoutes(api fiber.Router) {
	api.Get("/novels", controllers.GetNovels)
	api.Get("/novels/search", controllers.SearchNovels)
	api.Get("/novels/:id", controllers.GetNovel)
	api.Get("/novels/:novel_id/chapters", controllers.ListChapters)

	api.Post("/novels", middlewares.AuthMiddleware(), controllers.PostNovel)
	api.Put("/novels/:id", middlewares.AuthMiddleware(), controllers.UpdateNovel)
	api.Delete("/novels/:id", middlewares.AuthMiddleware(), controllers.DeleteNovel)
}
