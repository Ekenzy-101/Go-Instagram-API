package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ekenzy-101/Go-Gin-REST-API/app"
	"github.com/Ekenzy-101/Go-Gin-REST-API/handlers"
	"github.com/Ekenzy-101/Go-Gin-REST-API/models"
	"github.com/Ekenzy-101/Go-Gin-REST-API/routes"
	"github.com/Ekenzy-101/Go-Gin-REST-API/utils"
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
	ctx, cancel := beforeEachCreatePost()
	defer cancel()

	w := executeCreatePost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	subset := []string{"_id", "content", "category", "title", "userId", "createdAt", "updatedAt"}
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Subset(t, utils.GetMapKeys(body), subset)

	afterEachCreatePost(ctx)
}
func TestCreatePostFailsWithInvalidInputs(t *testing.T) {
	ctx, cancel := beforeEachCreatePost()
	defer cancel()

	category = ""
	content = ""
	title = ""
	w := executeCreatePost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	subset := []string{"content", "category", "title"}
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Subset(t, utils.GetMapKeys(body), subset)

	afterEachCreatePost(ctx)
}
func TestCreatePostFailsIfUserIsNotLoggedIn(t *testing.T) {
	ctx, cancel := beforeEachCreatePost()
	defer cancel()

	token = ""
	w := executeCreatePost()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, body, "message")

	afterEachCreatePost(ctx)
}

func afterEachCreatePost(ctx context.Context) {
	collection := app.GetCollectionHandle(handlers.PostCollection)
	collection.DeleteMany(ctx, bson.D{})
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

func beforeEachCreatePost() (context.Context, context.CancelFunc) {
	utils.LoadEnvVariables("../.env.test")
	ctx, cancel := app.CreateDataBaseConnection()
	category = "Test"
	content = "Test"
	title = "Test"
	user := &models.User{ID: primitive.NewObjectID(), Email: "test@gmail.com"}
	token, _ = user.GenerateToken()

	return ctx, cancel
}
