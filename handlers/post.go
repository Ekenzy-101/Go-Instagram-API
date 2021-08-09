package handlers

import (
	"context"
	"net/http"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreatePost(c *gin.Context) {
	var post *models.Post
	collection := services.GetMongoDBCollection(config.PostsCollection)

	user, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	if err := c.ShouldBindJSON(&post); err != nil {
		errors := err.(validator.ValidationErrors)
		messages := helpers.GenerateErrorMessages(errors)
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	userId, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	post.NormalizeFields(true)
	post.UserID = userId

	res, err := collection.InsertOne(context.TODO(), post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	post.ID = res.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, post)
}

func DeletePost(c *gin.Context) {
	var post *models.Post
	_id := c.Param("_id")
	collection := services.GetMongoDBCollection(config.PostsCollection)

	user, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	postId, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: postId},
	}
	err = collection.FindOne(context.TODO(), filter).Decode(&post)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if post.UserID.Hex() != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot delete this post"})
		return
	}

	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

func GetPost(c *gin.Context) {
	var post *models.Post
	_id := c.Param("_id")
	collection := services.GetMongoDBCollection(config.PostsCollection)

	postId, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	filter := bson.D{primitive.E{Key: "_id", Value: postId}}
	err = collection.FindOne(context.TODO(), filter).Decode(&post)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

func GetPosts(c *gin.Context) {
	collection := services.GetMongoDBCollection(config.PostsCollection)
	posts := []models.Post{}

	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	for cursor.Next(context.TODO()) {
		post := models.Post{}
		err := cursor.Decode(&post)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		posts = append(posts, post)
	}

	c.JSON(http.StatusOK, posts)
}

func GetUserPosts(c *gin.Context) {
	posts := []models.Post{}
	collection := services.GetMongoDBCollection(config.PostsCollection)

	user, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	userId, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	filter := bson.D{primitive.E{Key: "userId", Value: userId}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	for cursor.Next(context.TODO()) {
		post := models.Post{}
		err := cursor.Decode(&post)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		posts = append(posts, post)
	}

	c.JSON(http.StatusOK, posts)
}

func UpdatePost(c *gin.Context) {
	var body *models.Post
	var post *models.Post
	_id := c.Param("_id")
	collection := services.GetMongoDBCollection(config.PostsCollection)
	user, ok := c.MustGet("user").(*services.AccessTokenClaim)

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		errors := err.(validator.ValidationErrors)
		messages := helpers.GenerateErrorMessages(errors)
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	postId, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: postId},
	}

	err = collection.FindOne(context.TODO(), filter).Decode(&post)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if post.UserID.Hex() != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot update this post"})
		return
	}

	body.NormalizeFields(false)

	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "category", Value: body.Category},
			primitive.E{Key: "content", Value: body.Content},
			primitive.E{Key: "title", Value: body.Title},
		},
		},
	}

	_, err = collection.UpdateByID(context.TODO(), post.ID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	post.UpdateFields(body)

	c.JSON(http.StatusOK, post)
}
