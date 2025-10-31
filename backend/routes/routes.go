package routes

import "github.com/gofiber/fiber/v2"

func RoutesList(app *fiber.App) {
	api := app.Group("/api")
	authRoutes(api)
	novelRoutes(api)
	userRoutes(api)
	bookmarkRoutes(api)
	chapterRoutes(api)
	genreRoutes(api)
	tagsRoutes(api)
}