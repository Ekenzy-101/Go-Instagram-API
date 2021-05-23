package routes

import (
	"net/http"

	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/gin-gonic/gin"
)


func Authorizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No cookie found"})
			return
		}

		user, err := models.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}