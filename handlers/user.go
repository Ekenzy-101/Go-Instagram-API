package handlers

import (
	"context"
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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	findOneOptions := options.FindOne().SetProjection(bson.M{"password": 0, "email": 0, "gender": 0, "phoneNo": 0})
	result := models.FindUser(ctx, bson.M{"username": c.Param("username")}, findOneOptions)
	if result.User == nil {
		c.JSON(result.StatusCode, result.ResponseBody)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": result.User})
}

func GetUserHomePosts(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	findOneOptions := options.FindOne().SetProjection(bson.M{"followingCount": 1, "postsCount": 1})
	result := models.FindUser(ctx, bson.M{"_id": cliams.ID}, findOneOptions)
	if result.User == nil {
		c.JSON(result.StatusCode, result.ResponseBody)
		return
	}

	user := result.User
	if user.FollowingCount == 0 && user.PostsCount == 0 {
		c.JSON(http.StatusOK, gin.H{"posts": bson.A{}})
		return
	}

	var err error
	limitQueryValue := c.Query("limit")
	limit := uint64(config.CommonPaginationLength)
	if limitQueryValue != "" {
		limit, err = strconv.ParseUint(limitQueryValue, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a positive integer", limitQueryValue)})
			return
		}
	}

	skipQueryValue := c.Query("skip")
	skip := uint64(0)
	if skipQueryValue != "" {
		skip, err = strconv.ParseUint(skipQueryValue, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a positive integer", skipQueryValue)})
			return
		}
	}
	matchStage := bson.M{"$match": bson.M{"userId": cliams.ID, "following": bson.M{"$ne": bson.A{}}}}
	projectStage := bson.M{"$project": bson.M{"_id": 0, "following": 1, "userId": 1}}
	groupStage := bson.M{
		"$group": bson.M{
			"_id": "$userId",
			"following": bson.M{
				"$accumulator": bson.M{
					"init":           "function() {return []}",
					"accumulate":     "function(state = [], following) {return [...state, ...following]}",
					"accumulateArgs": bson.A{"$following"},
					"merge":          "function(state1, state2) {return [...state1, ...state2]}",
					"lang":           "js",
				},
			},
		},
	}
	postsLookupStage := bson.M{
		"$lookup": bson.M{
			"from": config.PostsCollection,
			"let":  bson.M{"following": "$following", "userId": "$_id"},
			"pipeline": bson.A{
				bson.M{
					"$match": bson.M{
						"$expr": bson.M{
							"$or": bson.A{
								bson.M{"$in": bson.A{"$userId", "$$following"}},
								bson.M{"$eq": bson.A{"$userId", "$$userId"}},
							},
						},
					},
				},
				bson.M{"$sort": bson.M{"createdAt": -1}},
				bson.M{"$skip": skip},
				bson.M{"$limit": limit},
			},
			"as": "posts",
		},
	}
	postsUnwindStage := bson.M{"$unwind": bson.M{"path": "$posts"}}
	replaceRootStage := bson.M{"$replaceRoot": bson.M{"newRoot": "$posts"}}
	postsPipeline := bson.A{matchStage, projectStage, groupStage, postsLookupStage, postsUnwindStage, replaceRootStage}

	usersLookupStage := bson.M{
		"$lookup": bson.M{
			"from": config.UsersCollection,
			"let":  bson.M{"userId": "$userId"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$userId"}}}},
				bson.M{"$project": bson.M{"username": 1, "image": 1}},
			},
			"as": "user",
		},
	}
	usersUnWindStage := bson.M{"$unwind": bson.M{"path": "$user"}}
	usersProjectStage := bson.M{"$project": bson.M{"userId": 0}}
	usersPipeline := bson.A{usersLookupStage, usersUnWindStage, usersProjectStage}

	pipeline := append(postsPipeline, usersPipeline...)
	userDetailsCollection := services.GetMongoDBCollection(config.UserDetailsCollection)
	cursor, err := userDetailsCollection.Aggregate(ctx, pipeline)
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

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func GetUserProfilePosts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findOneOptions := options.FindOne().SetProjection(bson.M{"postsCount": 1})
	findUserResult := models.FindUser(ctx, bson.M{"username": c.Param("username")}, findOneOptions)
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
		return
	}

	user := findUserResult.User

	var err error
	limitQueryValue := c.Query("limit")
	limit := uint64(config.CommonPaginationLength)
	if limitQueryValue != "" {
		limit, err = strconv.ParseUint(limitQueryValue, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid integer", limitQueryValue)})
			return
		}
	}

	skipQueryValue := c.Query("skip")
	skip := uint64(0)
	if skipQueryValue != "" {
		skip, err = strconv.ParseUint(skipQueryValue, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid integer", skipQueryValue)})
			return
		}
	}

	findOptions := options.Find().SetProjection(models.PostProjection).SetSort(bson.M{"createdAt": -1})
	findOptions = findOptions.SetSkip(int64(skip)).SetLimit(int64(limit))

	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	cursor, err := postsCollection.Find(ctx, bson.M{"userId": user.ID}, findOptions)
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

	hasNextPage := int(limit+skip) < user.PostsCount
	c.JSON(http.StatusOK, gin.H{"posts": posts, "hasNextPage": hasNextPage})
}

func GetUserSavedPosts(c *gin.Context) {
	cliams, ok := c.MustGet("user").(*services.AccessTokenClaim)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse decoded token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findUserResult := models.FindUser(ctx, bson.M{"_id": cliams.ID}, options.FindOne().SetProjection(bson.M{"_id": 1}))
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
		return
	}

	var err error
	limitQueryValue := c.Query("limit")
	limit := uint64(config.CommonPaginationLength)
	if limitQueryValue != "" {
		limit, err = strconv.ParseUint(limitQueryValue, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid integer", limitQueryValue)})
			return
		}
	}

	skipQueryValue := c.Query("skip")
	skip := uint64(0)
	if skipQueryValue != "" {
		skip, err = strconv.ParseUint(skipQueryValue, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid integer", skipQueryValue)})
			return
		}
	}

	documentsSkipped := skip / config.LargePaginationLength
	position := skip % config.LargePaginationLength

	matchStage := bson.M{"$match": bson.M{"userId": cliams.ID}}
	sortStage := bson.M{"$sort": bson.M{"createdAt": -1}}
	skipStage := bson.M{"$skip": documentsSkipped}
	limitStage := bson.M{"$limit": 1}
	projectStage := bson.M{
		"$project": bson.M{
			"_id":        0,
			"savedPosts": bson.M{"$slice": bson.A{"$savedPosts", position, limit}},
		},
	}
	lookupStage := bson.M{
		"$lookup": bson.M{
			"from": config.PostsCollection,
			"let":  bson.M{"savedPosts": "$savedPosts"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$savedPosts"}}}},
				bson.M{"$project": models.PostProjection},
			},
			"as": "savedPosts",
		},
	}
	unwindStage := bson.M{"$unwind": "$savedPosts"}
	replaceRootStage := bson.M{"$replaceRoot": bson.M{"newRoot": "$savedPosts"}}

	pipeline := bson.A{matchStage, sortStage, skipStage, limitStage, projectStage, lookupStage, unwindStage, replaceRootStage}

	userDetailsCollection := services.GetMongoDBCollection(config.UserDetailsCollection)
	cursor, err := userDetailsCollection.Aggregate(ctx, pipeline)
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

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func GetUserSimilarPosts(c *gin.Context) {
	postIdParamValue := c.Param("_id")
	postId, err := primitive.ObjectIDFromHex(postIdParamValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v is not a valid postId", postIdParamValue)})
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

	c.JSON(http.StatusOK, gin.H{"posts": users[0].Posts})
}

func GetUserTaggedPosts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	findUserResult := models.FindUser(ctx, bson.M{"username": c.Param("username")}, options.FindOne().SetProjection(bson.M{"_id": 1}))
	if findUserResult.User == nil {
		c.JSON(findUserResult.StatusCode, findUserResult.ResponseBody)
		return
	}

	var err error
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
		bson.M{
			"$match": bson.M{
				"caption": bson.M{"$regex": fmt.Sprintf("@%v", c.Param("username")), "$options": "m"},
			},
		},
		bson.M{"$project": models.PostProjection},
		bson.M{"$sort": bson.M{"createdAt": -1}},
		bson.M{"$skip": skip},
		bson.M{"$limit": limit},
	}
	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	cursor, err := postsCollection.Aggregate(ctx, pipeline)
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

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}
