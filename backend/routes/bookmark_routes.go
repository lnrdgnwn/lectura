package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func bookmarkRoutes(api fiber.Router) {
	api.Post("/bookmarks",  middlewares.AuthMiddleware(), controllers.AddBookmark)
	api.Delete("/bookmarks/:novel_id", middlewares.AuthMiddleware(), controllers.RemoveBookmark)
	api.Get("/users/me/bookmarks",  middlewares.AuthMiddleware(), controllers.ListBookmarks)
}
