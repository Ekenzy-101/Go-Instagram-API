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

func TestGetPostsSucceedsIfUserIsLoggedIn(t *testing.T) {
	beforeEachGetPosts()

	w := executeGetPosts()

	body := []map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	subset := []string{"_id", "content", "category", "title", "userId", "createdAt", "updatedAt"}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Subset(t, helpers.GetMapKeys(body[0]), subset)
	assert.Subset(t, helpers.GetMapKeys(body[1]), subset)

	afterEachGetPosts()
}

func TestGetPostsFailsIfUserIsNotLoggedIn(t *testing.T) {
	beforeEachGetPosts()

	token = ""
	w := executeGetPosts()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, body, "message")

	afterEachGetPosts()
}

func afterEachGetPosts() {
	collection := services.GetMongoDBCollection(config.PostsCollection)
	collection.DeleteMany(context.TODO(), bson.D{})
}

func executeGetPosts() *httptest.ResponseRecorder {
	var router = routes.SetupRouter()
	body := map[string]string{"content": content, "category": category, "title": title}
	jsonString, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts", bytes.NewReader(jsonString))
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	router.ServeHTTP(w, req)

	return w
}

func beforeEachGetPosts() {
	services.CreateMongoDBConnection()

	user := models.User{ID: primitive.NewObjectID(), Email: "test@gmail.com"}
	token, _ = user.GenerateToken()

	post1 := models.Post{ID: primitive.NewObjectID(), Category: "Test", Title: "Test", Content: "Test", UserID: user.ID}
	post2 := models.Post{ID: primitive.NewObjectID(), Category: "Test", Title: "Test", Content: "Test", UserID: user.ID}
	post1.NormalizeFields(true)
	post2.NormalizeFields(true)

	collection := services.GetMongoDBCollection(config.PostsCollection)
	collection.InsertMany(context.TODO(), []interface{}{post1, post2})
}
