package mocks

import (
	"context"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUserHomePosts() (string, error) {
	authUser := &models.User{
		Email:          "authuser@gmail.com",
		Username:       "authuser",
		PostsCount:     3,
		FollowingCount: 1,
	}
	authUser.NormalizeFields(true)

	userToFollow := &models.User{
		Email:      "usertofollow@gmail.com",
		Username:   "usertofollow",
		PostsCount: 3,
	}
	userToFollow.NormalizeFields(true)

	posts := bson.A{}
	for i := 0; i < 3; i++ {
		post := &models.Post{}
		post.NormalizeFields(userToFollow.ID)
		posts = append(posts, post)
	}

	for i := 0; i < 3; i++ {
		post := &models.Post{}
		post.NormalizeFields(authUser.ID)
		posts = append(posts, post)
	}

	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	_, err := postsCollection.InsertMany(context.Background(), posts)
	if err != nil {
		return "", err
	}

	authUserDetails := models.UserDetails{
		ID:        primitive.NewObjectID(),
		UserID:    authUser.ID,
		Following: bson.A{userToFollow.ID},
	}

	userDetailsCollection := services.GetMongoDBCollection(config.UserDetailsCollection)
	_, err = userDetailsCollection.InsertOne(context.Background(), authUserDetails)
	if err != nil {
		return "", err
	}

	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	_, err = usersCollection.InsertMany(context.Background(), bson.A{userToFollow, authUser})
	if err != nil {
		return "", err
	}

	return authUser.GenerateAccessToken()
}
