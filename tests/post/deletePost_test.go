package tests

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/mocks"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeletePostTestSuite struct {
	suite.Suite
	CommentsCollection *mongo.Collection
	InvalidID          string
	PostID             primitive.ObjectID
	PostsCollection    *mongo.Collection
	ResponseBody       bson.M
	RepliesCollection  *mongo.Collection
	Token              string
	UserID             primitive.ObjectID
	UsersCollection    *mongo.Collection
}

func (suite *DeletePostTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.CommentsCollection = services.GetMongoDBCollection(config.CommentsCollection)
	suite.RepliesCollection = services.GetMongoDBCollection(config.RepliesCollection)
	suite.PostsCollection = services.GetMongoDBCollection(config.PostsCollection)
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
}

func (suite *DeletePostTestSuite) SetupTest() {
	suite.ResponseBody = bson.M{}
	suite.InvalidID = ""

	result, err := mocks.DeletePost()
	if err != nil {
		log.Fatal(err)
	}

	suite.PostID = result.PostID
	suite.UserID = result.UserID
	suite.Token = result.Token
}

func (suite *DeletePostTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	request, err := http.NewRequest(http.MethodDelete, "/posts/"+suite.PostID.Hex()+suite.InvalidID, nil)
	if err != nil {
		return nil, err
	}

	request.AddCookie(&http.Cookie{Name: config.AccessTokenCookieName, Value: suite.Token})
	response := httptest.NewRecorder()
	router := routes.SetupRouter()
	router.ServeHTTP(response, request)
	err = json.NewDecoder(response.Body).Decode(&suite.ResponseBody)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (suite *DeletePostTestSuite) TearDownTest() {
	_, err := suite.CommentsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.PostsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.RepliesCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *DeletePostTestSuite) Test_DeletePost_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	err = suite.CommentsCollection.FindOne(context.Background(), bson.M{"postId": suite.PostID}).Err()
	suite.ErrorIs(err, mongo.ErrNoDocuments)

	err = suite.RepliesCollection.FindOne(context.Background(), bson.M{"postId": suite.PostID}).Err()
	suite.ErrorIs(err, mongo.ErrNoDocuments)

	err = suite.PostsCollection.FindOne(context.Background(), bson.M{"_id": suite.PostID}).Err()
	suite.ErrorIs(err, mongo.ErrNoDocuments)

	err = suite.UsersCollection.FindOne(context.Background(), bson.M{"_id": suite.UserID, "posts": bson.M{"$size": 0}}).Err()
	suite.NoError(err)

	suite.Equal(http.StatusOK, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *DeletePostTestSuite) Test_DeletePost_FailsIfPostIdIsInvalid() {
	suite.InvalidID = "invalid"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *DeletePostTestSuite) Test_DeletePost_FailsIfPostNotFound() {
	suite.PostID = primitive.NewObjectID()

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusNotFound, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *DeletePostTestSuite) Test_DeletePost_FailsIfUserNotFound() {
	_, err := suite.UsersCollection.DeleteOne(context.Background(), bson.M{"_id": suite.UserID})
	if err != nil {
		log.Fatal(err)
	}

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusNotFound, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *DeletePostTestSuite) Test_DeletePost_FailsIfUserNotLoggedIn() {
	suite.Token = ""

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusUnauthorized, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *DeletePostTestSuite) Test_DeletePost_FailsIfUserNotOwnerOfPost() {
	_, err := suite.PostsCollection.UpdateByID(context.Background(), suite.PostID, bson.M{"$set": bson.M{"userId": primitive.NewObjectID()}})
	if err != nil {
		log.Fatal(err)
	}

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusForbidden, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func TestDeletePostTestSuite(t *testing.T) {
	suite.Run(t, new(DeletePostTestSuite))
}
