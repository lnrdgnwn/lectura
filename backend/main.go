package main

import (
	"final_project/config"
	"final_project/database"
	"final_project/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	_ "final_project/docs"

	swagger "github.com/gofiber/swagger"
)

// @title           Final Project Novel API
// @version         1.0
// @description     REST API platform lECTURA (auth, users, novels, chapters, genres, tags, bookmarks).
// @contact.name    Your Name
// @contact.email   your.email@example.com
// @host            localhost:3000
// @BasePath        /api/v1
// @schemes         http
// @accept          json
// @produce         json
func main() {
	config.ENVLoad()
	database.DBLoad()
	database.DBMigrate()
	database.Redis()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	}))

	app.Static("/assets", "./assets")

	app.Get("/swagger/*", swagger.HandlerDefault)

	routes.RoutesList(app)

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}
