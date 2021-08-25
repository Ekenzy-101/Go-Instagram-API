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
	"github.com/Ekenzy-101/Go-Gin-REST-API/mocks"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SavePostTestSuite struct {
	suite.Suite
	InvalidID             string
	PostID                primitive.ObjectID
	PostsCollection       *mongo.Collection
	ResponseBody          bson.M
	Token                 string
	UserID                primitive.ObjectID
	UsersCollection       *mongo.Collection
	UserDetailsCollection *mongo.Collection
}

func (suite *SavePostTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.PostsCollection = services.GetMongoDBCollection(config.PostsCollection)
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
	suite.UserDetailsCollection = services.GetMongoDBCollection(config.UserDetailsCollection)
}

func (suite *SavePostTestSuite) SetupTest() {
	suite.ResponseBody = bson.M{}
	suite.InvalidID = ""

	result, err := mocks.SavePost()
	if err != nil {
		log.Fatal(err)
	}

	suite.PostID = result.PostID
	suite.UserID = result.UserID
	suite.Token = result.Token
}

func (suite *SavePostTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%v%v/save", suite.PostID.Hex(), suite.InvalidID), nil)
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

func (suite *SavePostTestSuite) TearDownTest() {
	_, err := suite.PostsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.UserDetailsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *SavePostTestSuite) Test_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	err = suite.UserDetailsCollection.FindOne(context.Background(), bson.M{"userId": suite.UserID, "savedPosts": suite.PostID}).Err()
	suite.NoError(err)

	suite.Equal(http.StatusOK, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *SavePostTestSuite) Test_FailsIfPostIdIsInvalid() {
	suite.InvalidID = "invalid"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusBadRequest, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *SavePostTestSuite) Test_FailsIfPostNotFound() {
	suite.PostID = primitive.NewObjectID()

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusNotFound, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *SavePostTestSuite) Test_FailsIfUserNotFound() {
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

func (suite *SavePostTestSuite) Test_FailsIfUserNotLoggedIn() {
	suite.Token = ""

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(http.StatusUnauthorized, response.Code)
	suite.Contains(suite.ResponseBody, "message")
}

func TestSavePostTestSuite(t *testing.T) {
	suite.Run(t, new(SavePostTestSuite))
}
