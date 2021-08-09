package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LoginRequestBody struct {
	Email    string `json:"email" binding:"email,max=255"`
	Password string `json:"password" binding:"required,min=6"`
}

func Register(c *gin.Context) {
	user := models.User{}
	messages := helpers.ValidateRequestBody(c, &user)
	if messages != nil {
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	user.NormalizeFields(true)
	err := user.HashPassword()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	collection := services.GetMongoDBCollection(config.UsersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err = collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User already exists"})
		return
	}

	if mongo.IsTimeout(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "We couldn't complete your request. Please try again"})
		return
	}

	if mongo.IsNetworkError(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Network Error " + err.Error()})
		return
	}

	token, err := user.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	user.Password = ""
	c.SetCookie("token", token, config.AccessTokenTTLInSeconds, "/", "", config.IsProduction, true)
	c.JSON(http.StatusCreated, user)
}

func Login(c *gin.Context) {
	requestBody := LoginRequestBody{}
	messages := helpers.ValidateRequestBody(c, &requestBody)
	if messages != nil {
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	user := &models.User{}
	filter := bson.M{"email": strings.ToLower(requestBody.Email)}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	collection := services.GetMongoDBCollection(config.UsersCollection)
	err := collection.FindOne(ctx, filter).Decode(user)

	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Email or Password"})
		return
	}

	if mongo.IsTimeout(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "We couldn't complete your request. Please try again"})
		return
	}

	if mongo.IsNetworkError(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Network Error " + err.Error()})
		return
	}

	matches, err := user.ComparePassword(requestBody.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if !matches {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Email or Password"})
		return
	}

	token, err := user.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	user.Password = ""
	c.SetCookie("token", token, config.AccessTokenTTLInSeconds, "/", "", config.IsProduction, true)
	c.JSON(http.StatusOK, user)
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
