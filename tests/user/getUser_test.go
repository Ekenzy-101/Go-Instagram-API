package tests

import (
	"context"
	"encoding/json"
	"fmt"
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
	"go.mongodb.org/mongo-driver/mongo"
)

type GetUserTestSuite struct {
	suite.Suite
	ResponseBody    bson.M
	Username        string
	UsersCollection *mongo.Collection
}

func (suite *GetUserTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
}

func (suite *GetUserTestSuite) SetupTest() {
	suite.ResponseBody = bson.M{}
	suite.Username = "testuser"

	user := models.User{
		Email:    "test@gmail.com",
		Gender:   "Male",
		PhoneNo:  "+23480000",
		Posts:    []bson.M{},
		Password: "123456",
		Username: suite.Username,
	}
	_, err := suite.UsersCollection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *GetUserTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%v", suite.Username), nil)
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

func (suite *GetUserTestSuite) TearDownTest() {
	_, err := suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *GetUserTestSuite) Test_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	subset := helpers.GetMapKeys(suite.ResponseBody["user"])
	list := helpers.GetStructFields(models.User{}, bson.A{"email", "password", "phoneNo", "gender"})

	suite.Equal(response.Code, http.StatusOK)
	suite.Subset(list, subset)
}

func (suite *GetUserTestSuite) Test_FailsIfUserNotFound() {
	suite.Username = "anotherusername"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(response.Code, http.StatusNotFound)
	suite.Contains(suite.ResponseBody, "message")
}

func TestGetUserTestSuite(t *testing.T) {
	suite.Run(t, new(GetUserTestSuite))
}
