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
	"github.com/Ekenzy-101/Go-Gin-REST-API/mocks"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreatePostResponseBody struct {
	Post       bson.M   `json:"post"`
	URLs       []string `json:"urls"`
	Message    string   `json:"message"`
	ImageCount string   `json:"imageCount"`
}

type CreatePostTestSuite struct {
	suite.Suite
	ImageCount      int
	ResponseBody    CreatePostResponseBody
	Token           string
	PostsCollection *mongo.Collection
	UsersCollection *mongo.Collection
}

func (suite *CreatePostTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.PostsCollection = services.GetMongoDBCollection(config.PostsCollection)
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
}

func (suite *CreatePostTestSuite) SetupTest() {
	suite.ImageCount = 2
	suite.ResponseBody = CreatePostResponseBody{}

	result, err := mocks.CreatePost()
	if err != nil {
		log.Fatal(err)
	}

	suite.Token = result.Token
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
	_, err := suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.PostsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *CreatePostTestSuite) Test_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.M{"email": "test@gmail.com", "posts.0.images": bson.M{"$size": suite.ImageCount}}
	err = suite.UsersCollection.FindOne(context.Background(), filter).Err()

	exclude := bson.A{"userId", "imageCount"}
	responseBody := suite.ResponseBody

	suite.NoError(err)
	suite.Equal(response.Code, http.StatusCreated)
	suite.Equal(len(responseBody.URLs), suite.ImageCount)
	suite.Subset(helpers.GetStructFields(models.Post{}, exclude), helpers.GetMapKeys(responseBody.Post))
	suite.Subset(bson.A{"", "testuser"}, helpers.GetMapValues(responseBody.Post["user"]))
}

func (suite *CreatePostTestSuite) Test_FailsWithInvalidInputs() {
	suite.ImageCount = 0

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	responseBody := suite.ResponseBody

	suite.Equal(response.Code, http.StatusBadRequest)
	suite.NotEmpty(responseBody.ImageCount)
}

func (suite *CreatePostTestSuite) Test_FailsIfUserIsNotLoggedIn() {
	suite.Token = ""

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	responseBody := suite.ResponseBody

	suite.Equal(response.Code, http.StatusUnauthorized)
	suite.NotEmpty(responseBody.Message)
}

func (suite *CreatePostTestSuite) Test_FailsIfUserNotFound() {
	var err error
	user := &models.User{ID: primitive.NewObjectID()}
	suite.Token, err = user.GenerateAccessToken()
	if err != nil {
		log.Fatal(err)
	}

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	responseBody := suite.ResponseBody

	suite.Equal(response.Code, http.StatusNotFound)
	suite.NotEmpty(responseBody.Message)
}

func TestCreatePostTestSuite(t *testing.T) {
	suite.Run(t, new(CreatePostTestSuite))
}
