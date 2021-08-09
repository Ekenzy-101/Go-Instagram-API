package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	postId string
)

func TestGetPostSucceedsIfUserIsLoggedIn(t *testing.T) {
	beforeEachGetPost()

	w := executeGetPost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	subset := []string{"_id", "content", "category", "title", "userId", "createdAt", "updatedAt"}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Subset(t, helpers.GetMapKeys(body), subset)

	afterEachGetPost()
}

func TestGetPostFailsIfUserIsNotLoggedIn(t *testing.T) {
	beforeEachGetPost()

	token = ""
	w := executeGetPost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, body, "message")

	afterEachGetPost()
}

func TestGetPostFailsIfPostIdIsInvalid(t *testing.T) {
	beforeEachGetPost()

	postId = "invalid"
	w := executeGetPost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, strings.ToLower(body["message"]), "invalid")

	afterEachGetPost()
}

func TestGetPostFailsIfPostNotFound(t *testing.T) {
	beforeEachGetPost()

	postId = primitive.NewObjectID().Hex()
	w := executeGetPost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, strings.ToLower(body["message"]), "not found")

	afterEachGetPost()
}

func afterEachGetPost() {
	collection := services.GetMongoDBCollection(config.UsersCollection)
	collection.DeleteMany(context.TODO(), bson.D{})
}

func executeGetPost() *httptest.ResponseRecorder {
	var router = routes.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/"+postId, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	router.ServeHTTP(w, req)

	return w
}

func beforeEachGetPost() {
	helpers.LoadEnvVariables("../.env.test")
	services.CreateMongoDBConnection()

	user := &models.User{ID: primitive.NewObjectID(), Email: "test@gmail.com"}
	token, _ = user.GenerateToken()

	post := &models.Post{ID: primitive.NewObjectID(), Category: "Test", Title: "Test", Content: "Test", UserID: user.ID}
	post.NormalizeFields(true)
	postId = post.ID.Hex()

	collection := services.GetMongoDBCollection(config.PostsCollection)
	collection.InsertOne(context.TODO(), post)
}
