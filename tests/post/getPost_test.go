package tests

import (
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

type GetPostTestSuite struct {
	suite.Suite
	InvalidID       string
	PostID          primitive.ObjectID
	PostsCollection *mongo.Collection
	ResponseBody    bson.M
	UserID          primitive.ObjectID
	UsersCollection *mongo.Collection
}

func (suite *GetPostTestSuite) SetupSuite() {
	services.CreateMongoDBConnection()
	suite.PostsCollection = services.GetMongoDBCollection(config.PostsCollection)
	suite.UsersCollection = services.GetMongoDBCollection(config.UsersCollection)
}

func (suite *GetPostTestSuite) SetupTest() {
	suite.PostID = primitive.NewObjectID()
	suite.UserID = primitive.NewObjectID()
	suite.ResponseBody = bson.M{}
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
	}
	_, err = suite.UsersCollection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *GetPostTestSuite) ExecuteRequest() (*httptest.ResponseRecorder, error) {
	request, err := http.NewRequest(http.MethodGet, "/posts/"+suite.PostID.Hex()+suite.InvalidID, nil)
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

func (suite *GetPostTestSuite) TearDownTest() {
	_, err := suite.PostsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	_, err = suite.UsersCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *GetPostTestSuite) Test_Succeeds() {
	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	subset := helpers.GetStructFields(models.Post{}, bson.A{"userId", "imageCount"})

	suite.Equal(response.Code, http.StatusOK)
	suite.Subset(helpers.GetMapKeys(suite.ResponseBody["post"]), subset)
}

func (suite *GetPostTestSuite) Test_FailsIfPostIdIsInvalid() {
	suite.InvalidID = "invalid"

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(response.Code, http.StatusBadRequest)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *GetPostTestSuite) Test_FailsIfPostNotFound() {
	suite.PostID = primitive.NewObjectID()

	response, err := suite.ExecuteRequest()
	if err != nil {
		log.Fatal(err)
	}

	suite.Equal(response.Code, http.StatusNotFound)
	suite.Contains(suite.ResponseBody, "message")
}

func (suite *GetPostTestSuite) Test_FailsIfUserNotFound() {
	_, err := suite.UsersCollection.DeleteOne(context.Background(), bson.M{"_id": suite.UserID})
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

func TestGetPostTestSuite(t *testing.T) {
	suite.Run(t, new(GetPostTestSuite))
}
