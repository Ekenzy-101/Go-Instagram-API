package handlers

import (
	"context"
	"errors"
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateReply(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	reply := &models.Reply{}
	messages := helpers.ValidateRequestBody(c, reply)
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

	err := reply.NormalizeFields(cliams.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	findOneOptions = options.FindOne().SetProjection(bson.M{"_id": 1})
	findPostResult := models.FindPost(ctx, bson.M{"_id": reply.PostID}, findOneOptions)
	if findPostResult.Post == nil {
		c.JSON(findPostResult.StatusCode, findPostResult.ResponseBody)
		return
	}

	findCommentResult := models.FindComment(ctx, bson.M{"_id": reply.ReplyToID})
	if findCommentResult.Comment == nil {
		c.JSON(findCommentResult.StatusCode, findCommentResult.ResponseBody)
		return
	}

	if findCommentResult.Comment.PostID != reply.PostID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "The comment's postId you are replying to does not match with the reply"})
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
		_, err := repliesCollection.InsertOne(sessCtx, reply)
		if err != nil {
			return nil, err
		}

		commentsCollection := services.GetMongoDBCollection(config.CommentsCollection)
		_, err = commentsCollection.UpdateByID(sessCtx, reply.ReplyToID, bson.M{"$inc": bson.M{"repliesCount": 1}})
		if err != nil {
			return nil, err
		}

		postsCollection := services.GetMongoDBCollection(config.PostsCollection)
		_, err = postsCollection.UpdateByID(sessCtx, reply.PostID, bson.M{"$inc": bson.M{"repliesCount": 1}})
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	reply.SetUser(findUserResult.User)
	c.JSON(http.StatusCreated, gin.H{"reply": reply})
}

func DeleteReply(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	replyIdParamValue := c.Param("_id")
	replyId, err := primitive.ObjectIDFromHex(replyIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid replyId", replyIdParamValue)})
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

	reply := &models.Reply{}
	repliesCollection := services.GetMongoDBCollection(config.RepliesCollection)
	err = repliesCollection.FindOne(ctx, bson.M{"_id": replyId}).Decode(reply)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Reply not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if reply.UserID != cliams.ID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not allowed to delete this reply"})
		return
	}

	session, err := services.GetMongoDBSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, err := repliesCollection.DeleteOne(sessCtx, bson.M{"_id": reply.ID})
		if err != nil {
			return nil, err
		}

		commentsCollection := services.GetMongoDBCollection(config.CommentsCollection)
		_, err = commentsCollection.UpdateByID(sessCtx, reply.ReplyToID, bson.M{"$inc": bson.M{"repliesCount": -1}})
		if err != nil {
			return nil, err
		}

		postsCollection := services.GetMongoDBCollection(config.PostsCollection)
		_, err = postsCollection.UpdateByID(sessCtx, reply.PostID, bson.M{"$inc": bson.M{"repliesCount": -1}})
		if err != nil {
			return nil, err
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

func GetReplies(c *gin.Context) {
	replyToIdQueryValue := c.Query("replyToId")
	replyToId, err := primitive.ObjectIDFromHex(replyToIdQueryValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid replyToId", replyToIdQueryValue)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findCommentResult := models.FindComment(ctx, bson.M{"_id": replyToId})
	if findCommentResult.Comment == nil {
		c.JSON(findCommentResult.StatusCode, findCommentResult.ResponseBody)
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

	pipeline := bson.A{
		bson.M{"$match": bson.M{"replyToId": replyToId}},
		bson.M{"$sort": bson.M{"createdAt": -1}},
		bson.M{"$skip": skip},
		bson.M{"$limit": limit},
		bson.M{
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
		},
		bson.M{"$unwind": bson.M{"path": "$user"}},
		bson.M{"$project": bson.M{"userId": 0}},
	}

	repliesCollection := services.GetMongoDBCollection(config.RepliesCollection)
	cursor, err := repliesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	replies := []models.Reply{}
	err = cursor.All(ctx, &replies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	hasNextPage := (limit + skip) < findCommentResult.Comment.RepliesCount
	c.JSON(http.StatusOK, gin.H{"replies": replies, "hasNextPage": hasNextPage})
}
