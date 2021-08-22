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

	userIdParamValue := c.Param("_id")
	userToFollowId, err := primitive.ObjectIDFromHex(userIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid userId", userIdParamValue)})
		return
	}

	if cliams.ID == userToFollowId {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot follow yourself"})
		return
	}

	filter := bson.M{"_id": bson.M{"$in": bson.A{userToFollowId, cliams.ID}}}
	findOptions := options.Find().SetProjection(bson.M{"followersCount": 1, "followingCount": 1})
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
		// Check if authUser has already followed userToFollow
		filter := bson.M{"userId": userToFollowId, "followers": cliams.ID}
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

		upsert := true
		operations = []mongo.WriteModel{
			&mongo.UpdateOneModel{
				Filter: bson.M{
					"userId":         userToFollowId,
					"followersCount": bson.M{"$lt": config.LargePaginationLength},
				},
				Upsert: &upsert,
				Update: bson.M{
					"$push":        bson.M{"followers": bson.M{"$each": bson.A{cliams.ID}, "$position": 0}},
					"$inc":         bson.M{"followersCount": 1},
					"$setOnInsert": models.NewUserDetails(bson.A{"followers", "followersCount"}),
				},
			},
			&mongo.UpdateOneModel{
				Filter: bson.M{
					"followingCount": bson.M{"$lt": config.LargePaginationLength},
					"userId":         cliams.ID,
				},
				Upsert: &upsert,
				Update: bson.M{
					"$push":        bson.M{"following": bson.M{"$each": bson.A{userToFollowId}, "$position": 0}},
					"$inc":         bson.M{"followingCount": 1},
					"$setOnInsert": models.NewUserDetails(bson.A{"following", "followingCount"}),
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
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Could not parse decoded token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userIdParamValue := c.Param("_id")
	userToUnfollowId, err := primitive.ObjectIDFromHex(userIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid userId", userIdParamValue)})
		return
	}

	if cliams.ID == userToUnfollowId {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You cannot unfollow yourself"})
		return
	}

	filter := bson.M{"_id": bson.M{"$in": bson.A{userToUnfollowId, cliams.ID}}}
	findOptions := options.Find().SetProjection(bson.M{"_id": 1})
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	cursor, err := usersCollection.Find(ctx, filter, findOptions)
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

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		operations := []mongo.WriteModel{
			&mongo.UpdateOneModel{
				Filter: bson.M{"userId": userToUnfollowId, "followers": cliams.ID},
				Update: bson.M{"$inc": bson.M{"followersCount": -1}, "$pull": bson.M{"followers": cliams.ID}},
			},
			&mongo.UpdateOneModel{
				Filter: bson.M{"userId": cliams.ID, "following": userToUnfollowId},
				Update: bson.M{"$inc": bson.M{"followingCount": -1}, "$pull": bson.M{"following": cliams.ID}},
			},
		}

		userDetailsCollection := services.GetMongoDBCollection(config.UserDetailsCollection)
		_, err = userDetailsCollection.BulkWrite(sessCtx, operations)
		if err != nil {
			return nil, err
		}

		operations = []mongo.WriteModel{
			&mongo.UpdateOneModel{
				Filter: bson.M{"_id": userToUnfollowId},
				Update: bson.M{"$inc": bson.M{"followersCount": -1}},
			},
			&mongo.UpdateOneModel{
				Filter: bson.M{"_id": cliams.ID},
				Update: bson.M{"$inc": bson.M{"followingCount": -1}},
			},
		}

		_, err = usersCollection.BulkWrite(sessCtx, operations)
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

func GetUserFollowers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userIdParamValue := c.Param("_id")
	userId, err := primitive.ObjectIDFromHex(userIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid userId", userIdParamValue)})
		return
	}

	findOneOptions := options.FindOne().SetProjection(bson.M{"followersCount": 1})
	findUserResult := models.FindUser(ctx, bson.M{"_id": userId}, findOneOptions)
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
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

	documentsSkipped := skip / config.LargePaginationLength
	position := skip % config.LargePaginationLength

	matchStage := bson.M{"$match": bson.M{"userId": userId}}
	sortStage := bson.M{"$sort": bson.M{"createdAt": -1}}
	skipStage := bson.M{"$skip": documentsSkipped}
	limitStage := bson.M{"$limit": 1}
	projectStage := bson.M{
		"$project": bson.M{
			"_id":       0,
			"followers": bson.M{"$slice": bson.A{"$followers", position, limit}},
		},
	}
	lookupStage := bson.M{
		"$lookup": bson.M{
			"from": config.UsersCollection,
			"let":  bson.M{"followers": "$followers"},
			"pipeline": bson.A{
				bson.M{
					"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$followers"}}},
				},
				bson.M{"$project": bson.M{"username": 1, "image": 1, "name": 1}},
			},
			"as": "followers",
		},
	}
	unwindStage := bson.M{"$unwind": "$followers"}
	replaceRootStage := bson.M{"$replaceRoot": bson.M{"newRoot": "$followers"}}
	pipeline := bson.A{matchStage, sortStage, skipStage, limitStage, projectStage, lookupStage, unwindStage, replaceRootStage}

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

	hasNextPage := (limit + skip) < findUserResult.User.FollowersCount
	c.JSON(http.StatusOK, gin.H{"users": users, "hasNextPage": hasNextPage})
}

func GetUserFollowing(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userIdParamValue := c.Param("_id")
	userId, err := primitive.ObjectIDFromHex(userIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid userId", userIdParamValue)})
		return
	}

	findOneOptions := options.FindOne().SetProjection(bson.M{"followingCount": 1})
	findUserResult := models.FindUser(ctx, bson.M{"_id": userId}, findOneOptions)
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
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

	documentsSkipped := skip / config.LargePaginationLength
	position := skip % config.LargePaginationLength

	matchStage := bson.M{"$match": bson.M{"userId": userId}}
	sortStage := bson.M{"$sort": bson.M{"createdAt": -1}}
	skipStage := bson.M{"$skip": documentsSkipped}
	limitStage := bson.M{"$limit": 1}
	projectStage := bson.M{
		"$project": bson.M{
			"_id":       0,
			"following": bson.M{"$slice": bson.A{"$following", position, limit}},
		},
	}
	lookupStage := bson.M{
		"$lookup": bson.M{
			"from": config.UsersCollection,
			"let":  bson.M{"following": "$following"},
			"pipeline": bson.A{
				bson.M{
					"$match": bson.M{
						"$expr": bson.M{"$in": bson.A{"$_id", "$$following"}},
					},
				},
				bson.M{"$project": bson.M{"username": 1, "image": 1, "name": 1}},
			},
			"as": "following",
		},
	}
	unwindStage := bson.M{"$unwind": "$following"}
	replaceRootStage := bson.M{"$replaceRoot": bson.M{"newRoot": "$following"}}
	pipeline := bson.A{matchStage, sortStage, skipStage, limitStage, projectStage, lookupStage, unwindStage, replaceRootStage}

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

	hasNextPage := (limit + skip) < findUserResult.User.FollowingCount
	c.JSON(http.StatusOK, gin.H{"users": users, "hasNextPage": hasNextPage})
}
