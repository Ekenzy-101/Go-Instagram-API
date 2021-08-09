package routes

import (
	"net/http"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-gonic/gin"
)

func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No cookie found"})
			return
		}

		option := services.JWTOption{
			Secret: config.AccessTokenSecret,
			Token:  token,
			Claims: &services.AccessTokenClaim{},
		}
		user, err := services.VerifyToken(option)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
