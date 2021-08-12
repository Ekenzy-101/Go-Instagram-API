package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type RegisterTestSuite struct {
	suite.Suite
	Email           string
	Name            string
	Password        string
	ResponseBody    bson.M
	Username        string
	UsersCollection *mongo.Collection
}

func (suite *RegisterTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
}

func (suite *RegisterTestSuite) SetupTest() {
	suite.Email = "test@gmail.com"
	suite.Password = "123456"
	suite.Name = "test"
	suite.Username = "testuser"
	suite.ResponseBody = bson.M{}
}

func (suite *RegisterTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	requestBodyMap := bson.M{
		"email":    suite.Email,
		"password": suite.Password,
		"name":     suite.Name,
		"username": suite.Username,
	}
	requestBodyBytes, err := json.Marshal(requestBodyMap)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(requestBodyBytes))
	if err != nil {
		return nil, err
	}

	response := httptest.NewRecorder()
	router := routes.SetupRouter()
	router.ServeHTTP(response, request)
	err = json.NewDecoder(response.Body).Decode(&suite.ResponseBody)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (suite *RegisterTestSuite) TearDownTest() {
	suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
}

func (suite *RegisterTestSuite) Test_Register_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	subset := []string{"_id", "name", "email", "username", "followerCount", "followingCount", "postCount"}

	suite.Equal(http.StatusCreated, response.Code)
	suite.Subset(helpers.GetMapKeys(suite.ResponseBody), subset)
	suite.Contains(response.Result().Header, "Set-Cookie")
}

func (suite *RegisterTestSuite) Test_Register_FailsWithInvalidInputs() {
	suite.Email = strings.Repeat("a", 247) + "@gmail.com"
	suite.Password = "111"
	suite.Name = ""
	suite.Username = ".username-"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	subset := []string{"password", "name", "email", "username"}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Subset(helpers.GetMapKeys(suite.ResponseBody), subset)
}

func (suite *RegisterTestSuite) Test_Register_FailsIfEmailExistsInDatabase() {
	user := models.User{
		Name:     suite.Name,
		Email:    suite.Email,
		Password: suite.Password,
		Username: "anotherusername",
	}
	collection := services.GetMongoDBCollection(config.UsersCollection)
	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Contains(suite.ResponseBody["message"], "email")
}
func (suite *RegisterTestSuite) Test_Register_FailsIfUsernameExistsInDatabase() {
	user := models.User{
		Name:     suite.Name,
		Email:    "anotheremail@gmail.com",
		Password: suite.Password,
		Username: suite.Username,
	}
	collection := services.GetMongoDBCollection(config.UsersCollection)
	collection.InsertOne(context.Background(), user)

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Contains(suite.ResponseBody["message"], "username")
}

func TestRegisterTestSuite(t *testing.T) {
	suite.Run(t, new(RegisterTestSuite))
}
