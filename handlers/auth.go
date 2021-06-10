package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Ekenzy-101/Go-Gin-REST-API/app"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LoginBody struct {
	Email    string `json:"email" binding:"email,max=255"`
	Password string `json:"password" binding:"required,min=6"`
}

const (
	UserCollection = "users"
)

func Register(c *gin.Context) {
	var user *models.User
	collection := app.GetCollectionHandle(UserCollection)

	err := c.ShouldBindJSON(&user)
	errors, ok := err.(validator.ValidationErrors)
	if ok {
		messages := utils.GenerateErrorMessages(errors)
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No body provided"})
		return
	}

	user.NormalizeFields(true)

	err = user.HashPassword()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	res, err := collection.InsertOne(context.TODO(), user)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User already exists"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	token, err := user.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.SetCookie("token", token, 3600, "/", "", false, true)
	c.JSON(http.StatusCreated, gin.H{"_id": res.InsertedID, "name": user.Name, "email": user.Email})
}

func Login(c *gin.Context) {
	var body *LoginBody
	collection := app.GetCollectionHandle(UserCollection)

	err := c.ShouldBindJSON(&body)
	// errors, ok := err.(validator.ValidationErrors)
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		messages := utils.GenerateErrorMessages(validationErrors)
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No body provided"})
		return
	}

	user := &models.User{}
	filter := bson.D{primitive.E{Key: "email", Value: strings.ToLower(body.Email)}}
	err = collection.FindOne(context.TODO(), filter).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Email or Password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if isSamePassword, err := user.ComparePassword(body.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	} else if !isSamePassword {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Email or Password"})
		return
	}

	token, err := user.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.SetCookie("token", token, 3600, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"_id": user.ID, "name": user.Name, "email": user.Email})
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, "Success")
}
