package handlers

import (
	"context"
	"fmt"
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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateComment(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	comment := &models.Comment{}
	messages := helpers.ValidateRequestBody(c, comment)
	if messages != nil {
		c.JSON(http.StatusBadRequest, messages)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findOneOptions := options.FindOne().SetProjection(bson.M{"username": 1, "image": 1})
	findUserResult := models.FindUser(ctx, bson.M{"_id": cliams.ID}, findOneOptions)
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
		return
	}

	comment.NormalizeFields(cliams.ID)
	update := bson.M{
		"$push": bson.M{"comments": bson.M{"$each": bson.A{comment}, "$position": 0, "$slice": config.CommonPaginationLength}},
		"$inc":  bson.M{"commentsCount": 1},
	}
	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	updateOneResult, err := postsCollection.UpdateByID(ctx, comment.PostID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if updateOneResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	commentsCollection := services.GetMongoDBCollection(config.CommentsCollection)
	_, err = commentsCollection.InsertOne(ctx, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	comment.SetUser(findUserResult.User)
	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

func DeleteComment(c *gin.Context) {
	// Validate decoded token
	// Validate commentId
	// Check if comment exist and authuser is owner
	// Find post for the given comment
	// Delete replies of comment
	// Delete comment in the subdocument and reduce count
	// Find top level comment and append to the post
}

func GetComments(c *gin.Context) {
	postIdQueryValue := c.Query("postId")
	postId, err := primitive.ObjectIDFromHex(postIdQueryValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid postId", postIdQueryValue)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findOneOptions := options.FindOne().SetProjection(bson.M{"commentsCount": 1})
	findPostResult := models.FindPost(ctx, bson.M{"_id": postId}, findOneOptions)
	if findPostResult.Post == nil {
		c.JSON(findPostResult.StatusCode, findPostResult.ResponseBody)
		return
	}

	limitQueryValue := c.Query("limit")
	limit := config.CommonPaginationLength
	if limitQueryValue != "" {
		limit, err = strconv.Atoi(limitQueryValue)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid integer", limitQueryValue)})
			return
		}
	}

	skipQueryValue := c.Query("skip")
	skip := 0
	if skipQueryValue != "" {
		skip, err = strconv.Atoi(skipQueryValue)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid integer", skipQueryValue)})
			return
		}
	}

	matchStage := bson.M{"$match": bson.M{"postId": postId}}
	sortStage := bson.M{"$sort": bson.M{"createdAt": -1}}
	skipStage := bson.M{"$skip": skip}
	limitStage := bson.M{"$limit": limit}
	lookupStage := bson.M{
		"$lookup": bson.M{
			"from": config.UsersCollection,
			"let":  bson.M{"userId": "$userId"},
			"pipeline": bson.A{
				bson.M{
					"$match": bson.M{
						"$expr": bson.M{"$eq": bson.A{"$_id", "$$userId"}},
					},
				},
				bson.M{
					"$project": bson.M{"image": 1, "username": 1, "_id": 0},
				},
			},
			"as": "user",
		},
	}
	unwindStage := bson.M{"$unwind": bson.M{"path": "$user"}}
	projectStage := bson.M{"$project": bson.M{"userId": 0}}

	pipeline := bson.A{matchStage, sortStage, skipStage, limitStage, lookupStage, unwindStage, projectStage}
	commentsCollection := services.GetMongoDBCollection(config.CommentsCollection)
	cursor, err := commentsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	comments := []models.Comment{}
	err = cursor.All(ctx, &comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	hasNextPage := (limit + skip) < findPostResult.Post.CommentsCount
	c.JSON(http.StatusOK, gin.H{"comments": comments, "hasNextPage": hasNextPage})
}
