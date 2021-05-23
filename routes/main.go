package routes

import (
	"github.com/Ekenzy-101/Go-Gin-REST-API/handlers"
	"github.com/gin-gonic/gin"
)


func SetupRouter() *gin.Engine {
	router := gin.Default()

	authRouter := router.Group("auth")
	{
		authRouter.POST("/login", handlers.Login)
		authRouter.POST("/logout", handlers.Logout)
		authRouter.POST("/register", handlers.Register)
	}

	postRouter := router.Group("posts")
	postRouter.Use(Authorizer())
	{
		postRouter.POST("", handlers.CreatePost)
		postRouter.DELETE("/:_id", handlers.DeletePost)
		postRouter.GET("/:_id", handlers.GetPost)
		postRouter.GET("", handlers.GetPosts)
		postRouter.GET("/me", handlers.GetUserPosts)
		postRouter.PUT("/:_id", handlers.UpdatePost)
	}

	return router
}