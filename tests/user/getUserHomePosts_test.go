package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/mocks"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetUserHomePostsTestSuite struct {
	suite.Suite
	ResponseBody          bson.M
	Limit                 string
	Skip                  string
	Username              string
	Token                 string
	PostsCollection       *mongo.Collection
	UsersCollection       *mongo.Collection
	UserDetailsCollection *mongo.Collection
}

func (suite *GetUserHomePostsTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
	suite.PostsCollection = services.GetMongoDBCollection(config.PostsCollection)
	suite.UserDetailsCollection = services.GetMongoDBCollection(config.UserDetailsCollection)
}

func (suite *GetUserHomePostsTestSuite) SetupTest() {
	suite.ResponseBody = bson.M{}
	suite.Username = "testuser"
	suite.Skip = "2"
	suite.Limit = "4"

	token, err := mocks.GetUserHomePosts()
	if err != nil {
		log.Fatal(err)
	}

	suite.Token = token
}

func (suite *GetUserHomePostsTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/me/posts/home?skip=%v&limit=%v", suite.Skip, suite.Limit), nil)
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

func (suite *GetUserHomePostsTestSuite) TearDownTest() {
	_, err := suite.PostsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.UserDetailsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *GetUserHomePostsTestSuite) Test_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	length, err := strconv.Atoi(suite.Limit)
	suite.NoError(err)

	suite.Equal(response.Code, http.StatusOK)
	suite.Len(suite.ResponseBody["posts"], length)
}

func (suite *GetUserHomePostsTestSuite) Test_FailsIfLimitQueryValueIsInvalid() {
	suite.Limit = "-1"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(response.Code, http.StatusBadRequest)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *GetUserHomePostsTestSuite) Test_FailsIfSkipQueryValueIsInvalid() {
	suite.Skip = "-2"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(response.Code, http.StatusBadRequest)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *GetUserHomePostsTestSuite) Test_FailsIfUserIsNotLoggedIn() {
	suite.Token = ""

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(response.Code, http.StatusUnauthorized)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *GetUserHomePostsTestSuite) Test_FailsIfUserNotFound() {
	var err error
	user := models.User{ID: primitive.NewObjectID()}
	suite.Token, err = user.GenerateAccessToken()
	if err != nil {
		log.Fatal(err)
	}

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(response.Code, http.StatusNotFound)
	suite.Contains(suite.ResponseBody, "message")
}

func TestGetUserHomePostsTestSuite(t *testing.T) {
	suite.Run(t, new(GetUserHomePostsTestSuite))
}
