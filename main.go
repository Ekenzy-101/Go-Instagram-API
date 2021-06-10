package main

import (
	"os"

	"github.com/Ekenzy-101/Go-Gin-REST-API/app"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/utils"
)

func main() {
	if os.Getenv("GIN_MODE") != "release" {
		utils.LoadEnvVariables()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	_, cancel := app.CreateDataBaseConnection()
	defer cancel()

	router := routes.SetupRouter()
	router.Run(":" + port)
}
