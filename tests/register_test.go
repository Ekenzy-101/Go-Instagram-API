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
)

var (
	email    string
	password string
)

type RegisterTestSuite struct {
	suite.Suite
	Email        string
	Password     string
	Name         string
	ResponseBody map[string]string
}

func (suite *RegisterTestSuite) SetupTest() {
	services.CreateMongoDBConnection()
	suite.Email = "test@gmail.com"
	suite.Password = "123456"
	suite.Name = "test"
	suite.ResponseBody = map[string]string{}
}

func (suite *RegisterTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	requestBodyMap := map[string]interface{}{"email": suite.Email, "password": suite.Password, "name": suite.Name}
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
	collection := services.GetMongoDBCollection(config.UsersCollection)
	collection.DeleteMany(context.TODO(), bson.D{})
}

func (suite *RegisterTestSuite) Test_Register_SucceedsWithValidInputs() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	subset := []string{"_id", "name", "email"}

	suite.Equal(http.StatusCreated, response.Code)
	suite.Subset(helpers.GetMapKeys(suite.ResponseBody), subset)
	suite.Contains(response.Result().Header, "Set-Cookie")
}

func (suite *RegisterTestSuite) Test_Register_FailsWithInvalidInputs() {
	suite.Email = strings.Repeat("a", 247) + "@gmail.com"
	suite.Password = "111"
	suite.Name = ""

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	subset := []string{"password", "name", "email"}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Subset(helpers.GetMapKeys(suite.ResponseBody), subset)
}

func (suite *RegisterTestSuite) Test_Register_FailsIfUserExistsInDatabase() {
	user := models.User{
		Name:     suite.Name,
		Email:    suite.Email,
		Password: suite.Password,
	}
	collection := services.GetMongoDBCollection(config.UsersCollection)
	collection.InsertOne(context.TODO(), user)

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func TestRegisterTestSuite(t *testing.T) {
	suite.Run(t, new(RegisterTestSuite))
}
