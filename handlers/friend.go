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

func FollowUser(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Could not parse decoded token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userToFollowId, err := primitive.ObjectIDFromHex(c.Param("_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid UserId"})
		return
	}

	if cliams.ID == userToFollowId {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot follow yourself"})
		return
	}

	filter := bson.M{"_id": bson.M{"$in": bson.A{userToFollowId, cliams.ID}}}
	findOptions := &options.FindOptions{
		Projection: bson.M{"followersCount": 1, "followingCount": 1},
	}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	cursor, err := usersCollection.Find(ctx, filter, findOptions)
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

	if len(users) != 2 {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	session, err := services.GetMongoDBSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer session.EndSession(ctx)

	userDetailsCollection := services.GetMongoDBCollection(config.UserDetailsCollection)
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Check if auth user has already followed the user
		filter := bson.M{"userId": userToFollowId, "followers": bson.M{"$in": bson.A{cliams.ID}}}
		err = userDetailsCollection.FindOne(sessCtx, filter).Err()
		if !errors.Is(err, mongo.ErrNoDocuments) && err != nil {
			return nil, err
		}

		if err == nil {
			return nil, nil
		}

		operations := []mongo.WriteModel{
			&mongo.UpdateOneModel{
				Filter: bson.M{"_id": userToFollowId},
				Update: bson.M{"$inc": bson.M{"followersCount": 1}},
			},
			&mongo.UpdateOneModel{
				Filter: bson.M{"_id": cliams.ID},
				Update: bson.M{"$inc": bson.M{"followingCount": 1}},
			},
		}
		_, err = usersCollection.BulkWrite(sessCtx, operations)
		if err != nil {
			return nil, err
		}

		authUser, userToFollow := models.User{}, models.User{}
		for _, user := range users {
			if user.ID == cliams.ID {
				authUser = user
			}

			if user.ID == userToFollowId {
				userToFollow = user
			}
		}
		upsert := true
		operations = []mongo.WriteModel{
			&mongo.UpdateOneModel{
				Filter: bson.M{
					"followersCount": bson.M{"$lt": config.FriendsPaginationLength},
					"userId":         userToFollowId,
				},
				Upsert: &upsert,
				Update: bson.M{
					"$push": bson.M{
						"followers": bson.M{"$each": bson.A{cliams.ID}, "$position": 0},
					},
					"$inc": bson.M{"followersCount": 1},
					"$setOnInsert": bson.M{
						"followersSkipped": userToFollow.FollowersCount,
						"following":        bson.A{},
						"followingCount":   0,
					},
				},
			},
			&mongo.UpdateOneModel{
				Filter: bson.M{
					"followingCount": bson.M{"$lt": config.FriendsPaginationLength},
					"userId":         cliams.ID,
				},
				Upsert: &upsert,
				Update: bson.M{
					"$push": bson.M{
						"following": bson.M{"$each": bson.A{userToFollowId}, "$position": 0},
					},
					"$inc": bson.M{"followingCount": 1},
					"$setOnInsert": bson.M{
						"followingSkipped": authUser.FollowingCount,
						"followers":        bson.A{},
						"followersCount":   0,
					},
				},
			},
		}
		_, err = userDetailsCollection.BulkWrite(sessCtx, operations)
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

func UnfollowUser(c *gin.Context) {

}

func GetUserFollowers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userId, err := primitive.ObjectIDFromHex(c.Param("_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid UserId"})
		return
	}

	user := &models.User{}
	findOneOptions := &options.FindOneOptions{
		Projection: bson.M{"followersCount": 1},
	}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	err = usersCollection.FindOne(ctx, bson.M{"_id": userId}, findOneOptions).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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

	followersSkipped := (skip / config.FriendsPaginationLength) * config.FriendsPaginationLength
	position := skip % config.FriendsPaginationLength

	matchStage := bson.M{
		"$match": bson.M{"userId": userId, "followingSkipped": followersSkipped},
	}
	projectStage := bson.M{
		"$project": bson.M{
			"_id":       0,
			"followers": bson.M{"$slice": bson.A{"$followers", position, limit}},
		},
	}
	lookupMatchStage := bson.M{
		"$match": bson.M{
			"$expr": bson.M{"$in": bson.A{"$_id", "$$followers"}},
		},
	}
	lookupProjectStage := bson.M{
		"$project": bson.M{"username": 1, "image": 1, "name": 1},
	}
	lookupStage := bson.M{
		"$lookup": bson.M{
			"from":     config.UsersCollection,
			"let":      bson.M{"followers": "$followers"},
			"pipeline": bson.A{lookupMatchStage, lookupProjectStage},
			"as":       "followers",
		},
	}
	unwindStage := bson.M{"$unwind": "$followers"}
	replaceRootStage := bson.M{"$replaceRoot": bson.M{"newRoot": "$followers"}}
	pipeline := bson.A{matchStage, projectStage, lookupStage, unwindStage, replaceRootStage}

	userDetailsCollection := services.GetMongoDBCollection(config.UserDetailsCollection)
	cursor, err := userDetailsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	users := []bson.M{}
	err = cursor.All(ctx, &users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	hasNextPage := (limit + skip) < user.FollowersCount
	c.JSON(http.StatusOK, gin.H{"users": users, "hasNextPage": hasNextPage})
}

func GetUserFollowing(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userId, err := primitive.ObjectIDFromHex(c.Param("_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid UserId"})
		return
	}

	user := &models.User{}
	findOneOptions := &options.FindOneOptions{
		Projection: bson.M{"followingCount": 1},
	}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	err = usersCollection.FindOne(ctx, bson.M{"_id": userId}, findOneOptions).Decode(user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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

	followingSkipped := (skip / config.FriendsPaginationLength) * config.FriendsPaginationLength
	position := skip % config.FriendsPaginationLength

	matchStage := bson.M{
		"$match": bson.M{"userId": userId, "followingSkipped": followingSkipped},
	}
	projectStage := bson.M{
		"$project": bson.M{
			"_id":       0,
			"following": bson.M{"$slice": bson.A{"$following", position, limit}},
		},
	}
	lookupMatchStage := bson.M{
		"$match": bson.M{
			"$expr": bson.M{"$in": bson.A{"$_id", "$$following"}},
		},
	}
	lookupProjectStage := bson.M{
		"$project": bson.M{"username": 1, "image": 1, "name": 1},
	}
	lookupStage := bson.M{
		"$lookup": bson.M{
			"from":     config.UsersCollection,
			"let":      bson.M{"following": "$following"},
			"pipeline": bson.A{lookupMatchStage, lookupProjectStage},
			"as":       "following",
		},
	}
	unwindStage := bson.M{"$unwind": "$following"}
	replaceRootStage := bson.M{"$replaceRoot": bson.M{"newRoot": "$following"}}
	pipeline := bson.A{matchStage, projectStage, lookupStage, unwindStage, replaceRootStage}

	userDetailsCollection := services.GetMongoDBCollection(config.UserDetailsCollection)
	cursor, err := userDetailsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	users := []bson.M{}
	err = cursor.All(ctx, &users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	hasNextPage := (limit + skip) < user.FollowingCount
	c.JSON(http.StatusOK, gin.H{"users": users, "hasNextPage": hasNextPage})
}
