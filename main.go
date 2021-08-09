package main

import (
	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
)

func main() {
	services.CreateMongoDBConnection()
	router := routes.SetupRouter()
	router.Run(":" + config.Port)
}
