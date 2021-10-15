package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/handlers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func SetupRouter() *gin.Engine {
	binding.Validator = &helpers.DefaultValidator{}
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.ClientOrigin},
		AllowMethods:     []string{"PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("%v operation is not supported for resource %v", c.Request.Method, c.Request.URL.Path)})
	})

	router.GET("/healthcheck", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		client := services.GetMongoDBClient()
		err := client.Ping(ctx, readpref.Primary())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	authRouter := router.Group("auth")
	{
		authRouter.POST("/login", handlers.Login)
		authRouter.POST("/logout", handlers.Logout)
		authRouter.POST("/register", handlers.Register)
	}

	commentRouter := router.Group("comments")
	{
		commentRouter.GET("", handlers.GetComments)
		commentRouter.POST("", Authorizer(true), handlers.CreateComment)
		commentRouter.DELETE("/:_id", Authorizer(true), handlers.DeleteComment)
	}

	friendshipRouter := router.Group("friendships")
	{
		friendshipRouter.GET("/:_id/followers", handlers.GetUserFollowers)
		friendshipRouter.GET("/:_id/following", handlers.GetUserFollowing)
		friendshipRouter.POST("/:_id/follow", Authorizer(true), handlers.FollowUser)
		friendshipRouter.POST("/:_id/unfollow", Authorizer(true), handlers.UnfollowUser)
	}

	postRouter := router.Group("posts")
	{
		postRouter.POST("", Authorizer(true), handlers.CreatePost)
		postRouter.POST("/:_id/save", Authorizer(true), handlers.SavePost)
		postRouter.DELETE("/:_id", Authorizer(true), handlers.DeletePost)
		postRouter.GET("/:_id", handlers.GetPost)
	}

	replyRouter := router.Group("replies")
	{
		replyRouter.GET("", handlers.GetReplies)
		replyRouter.POST("", Authorizer(true), handlers.CreateReply)
		replyRouter.DELETE("/:_id", Authorizer(true), handlers.DeleteReply)
	}

	userRouter := router.Group("users")
	{
		userRouter.GET("/:username", handlers.GetUser)
		userRouter.GET("/:username/posts/profile", handlers.GetUserProfilePosts)
		userRouter.GET("/:username/posts/:_id/similar", handlers.GetUserSimilarPosts)
		userRouter.GET("/:username/posts/tagged", handlers.GetUserTaggedPosts)
		userRouter.GET("/me/posts/home", Authorizer(true), handlers.GetUserHomePosts)
		userRouter.GET("/me/posts/saved", Authorizer(true), handlers.GetUserSavedPosts)
	}

	return router
}
