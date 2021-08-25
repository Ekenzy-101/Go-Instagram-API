package mocks

import (
	"context"
	"fmt"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostRouteMockResult struct {
	Token  string
	PostID primitive.ObjectID
	UserID primitive.ObjectID
}

func CreatePost() (*PostRouteMockResult, error) {
	documents := bson.A{}
	posts := []bson.M{}
	for i := 0; i < config.CommonPaginationLength; i++ {
		document := models.Post{Caption: "Test" + fmt.Sprint(i)}
		documents = append(documents, document)
		posts = append(posts, models.MapPostsToUserSubDocuments(document)...)
	}

	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	_, err := postsCollection.InsertMany(context.Background(), documents)
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:       primitive.NewObjectID(),
		Email:    "test@gmail.com",
		Username: "testuser",
		Posts:    posts,
	}

	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	_, err = usersCollection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}

	token, err := user.GenerateAccessToken()
	if err != nil {
		return nil, err
	}

	return &PostRouteMockResult{Token: token}, nil
}

func DeletePost() (*PostRouteMockResult, error) {
	userId := primitive.NewObjectID()
	post := models.Post{Caption: "Test", ID: primitive.NewObjectID(), UserID: userId}
	postsCollection := services.GetMongoDBCollection(config.PostsCollection)
	_, err := postsCollection.InsertOne(context.Background(), post)
	if err != nil {
		return nil, err
	}

	commentsCollection := services.GetMongoDBCollection(config.CommentsCollection)
	_, err = commentsCollection.InsertOne(context.Background(), models.Comment{PostID: post.ID})
	if err != nil {
		return nil, err
	}

	repliesCollection := services.GetMongoDBCollection(config.RepliesCollection)
	_, err = repliesCollection.InsertOne(context.Background(), models.Reply{PostID: post.ID})
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:       userId,
		Email:    "test@gmail.com",
		Username: "testuser",
		Posts:    models.MapPostsToUserSubDocuments(post),
	}
	usersCollection := services.GetMongoDBCollection(config.UsersCollection)
	_, err = usersCollection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}

	token, err := user.GenerateAccessToken()
	if err != nil {
		return nil, err
	}

	result := &PostRouteMockResult{
		Token:  token,
		UserID: user.ID,
		PostID: post.ID,
	}
	return result, nil
}
