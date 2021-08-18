package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	user := &models.User{}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	findOneOptions := &options.FindOneOptions{
		Projection: bson.M{"password": 0, "email": 0},
	}
	err := usersCollection.FindOne(ctx, bson.M{"username": c.Param("username")}, findOneOptions).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetUserHomePosts(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	user := &models.User{}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	err := usersCollection.FindOne(ctx, bson.M{"_id": cliams.ID}).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, "User not found")
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user.Posts)
}

func GetUserProfilePosts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	filter := bson.M{"username": c.Param("username")}
	findOneOptions := &options.FindOneOptions{
		Projection: bson.M{"username": 1, "image": 1},
	}
	user := &models.User{}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	err := usersCollection.FindOne(ctx, filter, findOneOptions).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	limitQueryValue := c.Query("limit")
	limit := int64(config.CommonPaginationLength)
	if limitQueryValue != "" {
		limit, err = strconv.ParseInt(limitQueryValue, 10, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid integer", limitQueryValue)})
			return
		}
	}

	skipQueryValue := c.Query("skip")
	skip := int64(0)
	if skipQueryValue != "" {
		skip, err = strconv.ParseInt(skipQueryValue, 10, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid integer", skipQueryValue)})
			return
		}
	}

	filter = bson.M{"userId": user.ID}
	findOptions := &options.FindOptions{
		Limit:      &limit,
		Projection: bson.M{"images": 1, "likesCount": 1, "commentsCount": 1, "createdAt": 1},
		Skip:       &skip,
		Sort:       bson.M{"createdAt": -1},
	}
	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	cursor, err := postsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	posts := []bson.M{}
	err = cursor.All(ctx, &posts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func GetUserSavedPosts(c *gin.Context) {

}

func GetUserSimilarPosts(c *gin.Context) {
	postId, err := primitive.ObjectIDFromHex(c.Param("_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid PostId"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	matchStage := bson.M{"$match": bson.M{"username": c.Param("username")}}
	firstProjectStage := bson.M{
		"$project": bson.M{
			"posts": bson.M{
				"$filter": bson.M{
					"input": "$posts",
					"as":    "post",
					"cond":  bson.M{"$ne": bson.A{"$$post._id", postId}},
				},
			},
		},
	}
	secondProjectStage := bson.M{
		"$project": bson.M{
			"posts": bson.M{
				"$slice": bson.A{"$posts", 9},
			},
		},
	}
	pipeline := bson.A{matchStage, firstProjectStage, secondProjectStage}

	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	cursor, err := usersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	users := []models.User{}
	err = cursor.All(ctx, &users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if len(users) != 1 {
		c.JSON(http.StatusNotFound, bson.M{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, users[0].Posts)
}

func GetUserTaggedPosts(c *gin.Context) {

}
