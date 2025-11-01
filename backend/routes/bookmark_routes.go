package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func bookmarkRoutes(api fiber.Router) {
	// tambahkan bookmark (user harus login)
	api.Post("/bookmarks",  middlewares.AuthMiddleware(), controllers.AddBookmark)

	// hapus bookmark untuk novel tertentu (user harus login)
	api.Delete("/bookmarks/:novel_id", middlewares.AuthMiddleware(), controllers.RemoveBookmark)

	// daftar bookmark user yang sedang login
	api.Get("/users/me/bookmarks",  middlewares.AuthMiddleware(), controllers.ListBookmarks)
}
