package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func bookmarkRoutes(api fiber.Router) {
	api.Get("/bookmarks",  middlewares.AuthMiddleware(), controllers.ListBookmarks)
	api.Get("/bookmarks/novels", middlewares.AuthMiddleware(), controllers.GetNovelByBookmarks)
	api.Post("/bookmarks",  middlewares.AuthMiddleware(), controllers.AddBookmark)
	api.Delete("/bookmarks/:id", middlewares.AuthMiddleware(), controllers.RemoveBookmark)
}
