package routes

import (
	"net/http"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-gonic/gin"
)

func Authorizer(credentialsRequired bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie(config.AccessTokenCookieName)
		if err != nil && credentialsRequired {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No token found"})
			return
		}

		if err != nil {
			c.Next()
			return
		}

		option := services.JWTOption{
			Secret: config.AccessTokenSecret,
			Token:  accessToken,
			Claims: &services.AccessTokenClaim{},
		}
		user, err := services.VerifyToken(option)
		if err != nil && credentialsRequired {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}

		if err != nil {
			c.Next()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
