package routes

import (
	"final_project/controllers"
	"final_project/middlewares"

	"github.com/gofiber/fiber/v2"
)

func bookmarkRoutes(api fiber.Router) {
	// tambahkan bookmark (user harus login)
	api.Post("/bookmarks", controllers.AddBookmark, middlewares.AuthMiddleware())

	// hapus bookmark untuk novel tertentu (user harus login)
	api.Delete("/bookmarks/:novel_id", controllers.RemoveBookmark, middlewares.AuthMiddleware())

	// daftar bookmark user yang sedang login
	api.Get("/users/me/bookmarks", controllers.ListBookmarks, middlewares.AuthMiddleware())
}
