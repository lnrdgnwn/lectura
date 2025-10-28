package main

import (
	"final_project/config"
	"final_project/database"
	"final_project/routes"

	"fmt"
	"log"
	"net/http"
)

func main() {
	config.ENVLoad()
	database.InitDB()
	fmt.Println("Database Connected")
	defer database.DB.Close()

	database.Migrate()
	fmt.Println("Migration Success")

	routes.SetupRoutes()

	fmt.Println("Server running at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}