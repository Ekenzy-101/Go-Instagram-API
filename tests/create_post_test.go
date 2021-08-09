package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	category string
	content  string
	title    string
	token    string
)

func TestCreatePostSucceedsWithValidInputs(t *testing.T) {
	beforeEachCreatePost()

	w := executeCreatePost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	subset := []string{"_id", "content", "category", "title", "userId", "createdAt", "updatedAt"}
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Subset(t, helpers.GetMapKeys(body), subset)

	afterEachCreatePost()
}
func TestCreatePostFailsWithInvalidInputs(t *testing.T) {
	beforeEachCreatePost()

	category = ""
	content = ""
	title = ""
	w := executeCreatePost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	subset := []string{"category", "title"}
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Subset(t, helpers.GetMapKeys(body), subset)

	afterEachCreatePost()
}
func TestCreatePostFailsIfUserIsNotLoggedIn(t *testing.T) {
	beforeEachCreatePost()

	token = ""
	w := executeCreatePost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, body, "message")

	afterEachCreatePost()
}

func afterEachCreatePost() {
	collection := services.GetMongoDBCollection(config.PostsCollection)
	collection.DeleteMany(context.TODO(), bson.D{})
}

func executeCreatePost() *httptest.ResponseRecorder {
	var router = routes.SetupRouter()
	body := map[string]string{"content": content, "category": category, "title": title}
	jsonString, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewReader(jsonString))
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	router.ServeHTTP(w, req)

	return w
}

func beforeEachCreatePost() {
	helpers.LoadEnvVariables("../.env.test")
	services.CreateMongoDBConnection()
	category = "Test"
	content = "Test"
	title = "Test"
	user := &models.User{ID: primitive.NewObjectID(), Email: "test@gmail.com"}
	token, _ = user.GenerateToken()

}
