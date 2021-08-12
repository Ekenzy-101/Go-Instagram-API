package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreatePostResponseBody struct {
	Post       *models.Post `json:"post"`
	URLs       []string     `json:"urls"`
	Message    string       `json:"message"`
	ImageCount string       `json:"imageCount"`
}

type CreatePostTestSuite struct {
	suite.Suite
	ImageCount      int
	ResponseBody    CreatePostResponseBody
	Token           string
	PostsCollection *mongo.Collection
	UserID          primitive.ObjectID
	UsersCollection *mongo.Collection
}

func (suite *CreatePostTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.PostsCollection = services.GetMongoDBCollection(config.PostsCollection)
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
}

func (suite *CreatePostTestSuite) SetupTest() {
	suite.ImageCount = 2
	suite.UserID = primitive.NewObjectID()
	suite.ResponseBody = CreatePostResponseBody{}

	user := models.User{
		ID:       suite.UserID,
		Email:    "test@gmail.com",
		Username: "testuser",
		Posts:    []bson.M{},
	}
	_, err := suite.UsersCollection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}

	suite.Token, err = user.GenerateAccessToken()
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *CreatePostTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	body := bson.M{"caption": "Test", "location": "Test", "imageCount": suite.ImageCount}
	jsonString, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, "/posts", bytes.NewReader(jsonString))
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

func (suite *CreatePostTestSuite) TearDownTest() {
	suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
	suite.PostsCollection.DeleteMany(context.Background(), bson.M{})
}

func (suite *CreatePostTestSuite) Test_CreatePost_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	responseBody := suite.ResponseBody
	list := []interface{}{"", "testuser"}

	suite.Equal(response.Code, http.StatusCreated)
	suite.Equal(len(responseBody.URLs), suite.ImageCount)
	suite.Subset(list, helpers.GetMapValues(responseBody.Post.User))
}

func (suite *CreatePostTestSuite) Test_CreatePost_FailsWithInvalidInputs() {
	suite.ImageCount = 0

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	responseBody := suite.ResponseBody

	suite.Equal(response.Code, http.StatusBadRequest)
	suite.NotEmpty(responseBody.ImageCount)
}

func (suite *CreatePostTestSuite) Test_CreatePost_FailsIfUserIsNotLoggedIn() {
	suite.Token = ""

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	responseBody := suite.ResponseBody

	suite.Equal(response.Code, http.StatusUnauthorized)
	suite.NotEmpty(responseBody.Message)
}
func (suite *CreatePostTestSuite) Test_CreatePost_FailsIfUserNotFound() {
	deleteResult, err := suite.UsersCollection.DeleteOne(context.Background(), bson.M{"_id": suite.UserID})
	if err != nil {
		log.Fatal(err)
	}

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	responseBody := suite.ResponseBody

	suite.Equal(response.Code, http.StatusNotFound)
	suite.NotEmpty(deleteResult.DeletedCount)
	suite.NotEmpty(responseBody.Message)
}

func TestCreatePostTestSuite(t *testing.T) {
	suite.Run(t, new(CreatePostTestSuite))
}
