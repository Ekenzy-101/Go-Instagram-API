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

type LoginTestSuite struct {
	suite.Suite
	Email           string
	Password        string
	ResponseBody    bson.M
	UsersCollection *mongo.Collection
}

func (suite *LoginTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
}

func (suite *LoginTestSuite) SetupTest() {
	suite.Email = "test@gmail.com"
	suite.Password = "123456"
	suite.ResponseBody = bson.M{}

	user := models.User{
		Name:     "test",
		Email:    suite.Email,
		Password: suite.Password,
	}
	user.HashPassword()

	_, err := suite.UsersCollection.InsertOne(context.Background(), &user)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *LoginTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	requestBodyMap := bson.M{"email": suite.Email, "password": suite.Password}
	requestBodyBytes, err := json.Marshal(requestBodyMap)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(requestBodyBytes))
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

func (suite *LoginTestSuite) TearDownTest() {
	_, err := suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *LoginTestSuite) Test_Login_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	subset := []string{"_id", "name", "email"}

	suite.Equal(http.StatusOK, response.Code)
	suite.Subset(helpers.GetMapKeys(suite.ResponseBody), subset)
	suite.Contains(response.Result().Header, "Set-Cookie")
}

func (suite *LoginTestSuite) Test_Login_FailsWithInvalidInputs() {
	suite.Email = strings.Join(make([]string, 247), "a") + "@gmail.com"
	suite.Password = "111"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	subset := []string{"email", "password"}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Subset(helpers.GetMapKeys(suite.ResponseBody), subset)
}

func (suite *LoginTestSuite) Test_Login_FailsIfUserDoesNotExistInDatabase() {
	suite.Email = "doesnotexist@gmail.com"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *LoginTestSuite) Test_Login_FailsIfPasswordDoesNotMatch() {
	suite.Password = "notmatch"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func TestLoginTestSuite(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}
