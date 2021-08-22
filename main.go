package main

import (
	"log"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
)

func main() {
	services.CreateMongoDBConnection()
	router := routes.SetupRouter()
	err := router.Run(":" + config.Port)
	if err != nil {
		log.Fatal(err)
	}
}
