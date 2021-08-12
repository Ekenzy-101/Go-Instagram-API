package tests

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeletePostTestSuite struct {
	suite.Suite
	InvalidID       string
	PostID          primitive.ObjectID
	PostsCollection *mongo.Collection
	ResponseBody    bson.M
	Token           string
	UserID          primitive.ObjectID
	UsersCollection *mongo.Collection
}

func (suite *DeletePostTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.PostsCollection = services.GetMongoDBCollection(config.PostsCollection)
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
}

func (suite *DeletePostTestSuite) SetupTest() {
	suite.ResponseBody = bson.M{}
	suite.PostID = primitive.NewObjectID()
	suite.UserID = primitive.NewObjectID()
	suite.InvalidID = ""

	post := models.Post{
		ID:       suite.PostID,
		Location: "Test",
		Caption:  "Test",
		Images:   []string{},
		UserID:   suite.UserID,
	}
	_, err := suite.PostsCollection.InsertOne(context.Background(), post)
	if err != nil {
		log.Fatal(err)
	}

	user := models.User{
		ID:       suite.UserID,
		Email:    "test@gmail.com",
		Username: "testuser",
		Posts: []bson.M{{
			"_id": suite.PostID,
		}},
	}
	_, err = suite.UsersCollection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}

	suite.Token, err = user.GenerateAccessToken()
	if err != nil {
		log.Fatal(err)
	}
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
	suite.PostsCollection.DeleteMany(context.Background(), bson.M{})
	suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
}

func (suite *DeletePostTestSuite) Test_DeletePost_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

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
