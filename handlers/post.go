package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreatePost(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Could not parse decoded token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	user := &models.User{}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	err := usersCollection.FindOne(ctx, bson.M{"_id": cliams.ID}).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	post := &models.Post{}
	messages := helpers.ValidateRequestBody(c, post)
	if messages != nil {
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	post.NormalizeFields(user)
	keys := post.GeneratePresignedURLKeys()
	urls, err := services.GeneratePresignedURLs(keys)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	post.SetImages(urls)
	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	_, err = postsCollection.InsertOne(ctx, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	postDocuments := models.MapPostsToUserSubDocuments(*post)
	update := bson.M{
		"$push": bson.M{"posts": bson.M{"$each": postDocuments, "$position": 0, "$slice": config.PostsLengthInUserDocument}},
		"$inc":  bson.M{"postCount": 1},
	}
	_, err = usersCollection.UpdateByID(ctx, user.ID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	post.SetUser(user)
	c.JSON(http.StatusCreated, gin.H{"post": post, "urls": urls})
}

func DeletePost(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	user := &models.User{}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	opts := &options.FindOneOptions{
		Projection: bson.M{"posts": 1, "postCount": 1},
	}
	err := usersCollection.FindOne(ctx, bson.M{"_id": cliams.ID}, opts).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	postId, err := primitive.ObjectIDFromHex(c.Param("_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid postId"})
		return
	}

	post := &models.Post{}
	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	filter := bson.M{"_id": postId}
	opts = &options.FindOneOptions{
		Projection: bson.M{"userId": 1},
	}
	err = postsCollection.FindOne(ctx, filter, opts).Decode(post)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if post.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot delete this post"})
		return
	}

	deleteResult, err := postsCollection.DeleteOne(ctx, filter)
	if deleteResult.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	update := bson.M{
		"$pull": bson.M{"posts": bson.M{"_id": postId}},
		"$inc":  bson.M{"postCount": -1},
	}
	_, err = usersCollection.UpdateByID(ctx, user.ID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if user.PostCount > 12 {
		recentPost := bson.M{}
		postIds := user.GetPostIds()
		filter = bson.M{"_id": bson.M{"$nin": postIds}}
		opts = &options.FindOneOptions{
			Projection: bson.M{"images": 1, "likeCount": 1, "commentCount": 1},
			Sort:       bson.M{"createdAt": -1},
		}
		err := postsCollection.FindOne(ctx, filter, opts).Decode(&recentPost)
		if !errors.Is(err, mongo.ErrNoDocuments) && err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		update = bson.M{"$push": bson.M{"posts": bson.M{"$each": bson.A{recentPost}, "$sort": bson.M{"createdAt": -1}}}}
		_, err = usersCollection.UpdateByID(ctx, user.ID, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func GetPost(c *gin.Context) {
	postId, err := primitive.ObjectIDFromHex(c.Param("_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid postId"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	post := &models.Post{}
	collection := services.GetMongoDBCollection(config.PostsCollection)
	err = collection.FindOne(ctx, bson.M{"_id": postId}).Decode(post)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	user := &models.User{}
	collection = services.GetMongoDBCollection(config.UsersCollection)
	err = collection.FindOne(ctx, bson.M{"_id": post.UserID}).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	post.SetUser(user)
	c.JSON(http.StatusOK, post)
}

func GetPosts(c *gin.Context) {
	username := c.Query("username")

	switch {
	case c.Query("postId") != "":
		GetUserSimilarPosts(c)
	case username != "":
		GetUserProfilePosts(c)
	default:
		GetUserHomePosts(c)
	}

}

func GetUserHomePosts(c *gin.Context) {
	value, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Please provide a valid token"})
		return
	}

	cliams, ok := value.(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Please provide a valid token"})
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

	c.JSON(200, "ehll")
}

func GetUserSimilarPosts(c *gin.Context) {
	c.JSON(200, "ehll")
}

func GetUserProfilePosts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	filter := bson.M{"username": c.Query("username")}
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
	limit := int64(config.PostsLengthInUserDocument)
	if limitQueryValue != "" {
		limit, err = strconv.ParseInt(limitQueryValue, 10, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	skipQueryValue := c.Query("skip")
	skip := int64(0)
	if skipQueryValue != "" {
		skip, err = strconv.ParseInt(skipQueryValue, 10, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	filter = bson.M{"userId": user.ID}
	findOptions := &options.FindOptions{
		Limit:      &limit,
		Projection: bson.M{"images": 1, "likeCount": 1, "commentCount": 1, "createdAt": 1},
		Skip:       &skip,
		Sort:       bson.M{"createdAt": -1},
	}
	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	cursor, err := postsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	posts := []models.Post{}
	err = cursor.All(ctx, &posts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	postDocuments := models.MapPostsToUserSubDocuments(posts...)
	c.JSON(http.StatusOK, postDocuments)
}
