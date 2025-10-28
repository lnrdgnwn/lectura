package main

import (
	"final_project/config"
	"final_project/database"
	"final_project/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config.ENVLoad()
	database.DBLoad()
	database.DBMigrate()
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173/",
		AllowCredentials: true,
	}))
	app.Static("/assets", "./assets")

	routes.RoutesList(app)
	app.Listen(":3000")
}
