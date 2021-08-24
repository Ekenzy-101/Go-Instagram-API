package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

	post := &models.Post{}
	messages := helpers.ValidateRequestBody(c, post)
	if messages != nil {
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findUserResult := models.FindUser(ctx, bson.M{"_id": cliams.ID})
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.StatusCode)
		return
	}

	user := findUserResult.User
	post.NormalizeFields(user.ID)

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
		"$push": bson.M{"posts": bson.M{"$each": postDocuments, "$position": 0, "$slice": config.CommonPaginationLength}},
		"$inc":  bson.M{"postsCount": 1},
	}

	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
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

	postIdParamValue := c.Param("_id")
	postId, err := primitive.ObjectIDFromHex(postIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid postId", postIdParamValue)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findOneOptions := options.FindOne().SetProjection(bson.M{"posts": 1, "postsCount": 1})
	findUserResult := models.FindUser(ctx, bson.M{"_id": cliams.ID}, findOneOptions)
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
		return
	}

	findOneOptions = options.FindOne().SetProjection(bson.M{"userId": 1})
	findPostResult := models.FindPost(ctx, bson.M{"_id": postId}, findOneOptions)
	if findPostResult.Post == nil {
		c.JSON(findPostResult.StatusCode, findPostResult.ResponseBody)
		return
	}

	if findPostResult.Post.UserID != cliams.ID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot delete this post"})
		return
	}

	session, err := services.GetMongoDBSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		repliesCollection := services.GetMongoDBCollection(config.RepliesCollection)
		_, err := repliesCollection.DeleteMany(sessCtx, bson.M{"postId": postId})
		if err != nil {
			return nil, err
		}

		commentsCollection := services.GetMongoDBCollection(config.CommentsCollection)
		_, err = commentsCollection.DeleteMany(sessCtx, bson.M{"postId": postId})
		if err != nil {
			return nil, err
		}

		postsCollection := services.GetMongoDBCollection(config.PostsCollection)
		_, err = postsCollection.DeleteOne(sessCtx, bson.M{"_id": postId})
		if err != nil {
			return nil, err
		}

		user := findUserResult.User
		update := bson.M{
			"$pull": bson.M{"posts": bson.M{"_id": postId}},
			"$inc":  bson.M{"postsCount": -1},
		}
		usersCollection := services.GetMongoDBCollection(config.UsersCollection)
		_, err = usersCollection.UpdateByID(sessCtx, user.ID, update)
		if err != nil {
			return nil, err
		}

		if user.PostsCount > config.CommonPaginationLength {
			recentPost := bson.M{}
			filter := bson.M{"_id": bson.M{"$nin": user.GetPostIds()}, "userId": user.ID}
			findOneOptions := options.FindOne().SetProjection(models.PostProjection)
			findOneOptions.SetSort(bson.M{"createdAt": -1})

			err := postsCollection.FindOne(sessCtx, filter, findOneOptions).Decode(&recentPost)
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, nil
			}

			if err != nil {
				return nil, err
			}

			update = bson.M{"$push": bson.M{"posts": recentPost}}
			_, err = usersCollection.UpdateByID(sessCtx, user.ID, update)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func GetPost(c *gin.Context) {
	postIdParamValue := c.Param("_id")
	postId, err := primitive.ObjectIDFromHex(postIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid postId", postIdParamValue)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findPostResult := models.FindPost(ctx, bson.M{"_id": postId})
	if findPostResult.Post == nil {
		c.JSON(findPostResult.StatusCode, findPostResult.ResponseBody)
		return
	}

	post := findPostResult.Post

	findUserResult := models.FindUser(ctx, bson.M{"_id": post.UserID})
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
		return
	}

	post.SetUser(findUserResult.User)
	c.JSON(http.StatusOK, bson.M{"post": post})
}

func SavePost(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Could not parse decoded token"})
		return
	}

	postIdParamValue := c.Param("_id")
	postId, err := primitive.ObjectIDFromHex(postIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid postId", postIdParamValue)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findOneOptions := options.FindOne().SetProjection(bson.M{"_id": 1})
	findUserResult := models.FindUser(ctx, bson.M{"_id": cliams.ID}, findOneOptions)
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
		return
	}

	findPostResult := models.FindPost(ctx, bson.M{"_id": postId})
	if findPostResult.Post == nil {
		c.JSON(findPostResult.StatusCode, findPostResult.ResponseBody)
		return
	}

	update := bson.M{
		"$push":        bson.M{"savedPosts": bson.M{"$each": bson.A{postId}, "$position": 0}},
		"$inc":         bson.M{"savedPostsCount": 1},
		"$setOnInsert": models.NewUserDetails(bson.A{"savedPosts", "savedPostsCount"}),
	}
	filter := bson.M{"userId": cliams.ID, "savedPostsCount": bson.M{"$lt": config.LargePaginationLength}}

	userDetailsCollection := services.GetMongoDBCollection(config.UserDetailsCollection)
	_, err = userDetailsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
