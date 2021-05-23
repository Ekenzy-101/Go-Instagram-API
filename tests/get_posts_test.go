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

func TestGetPostsSucceedsIfUserIsLoggedIn(t *testing.T)  {
	ctx, cancel := beforeEachGetPosts()
	defer cancel()

	w := executeGetPosts()
	
	body := []map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	subset := []string{"_id", "content", "category", "title", "userId", "createdAt", "updatedAt"}
	assert.Equal(t,http.StatusOK, w.Code)
	assert.Subset(t, utils.GetMapKeys(body[0]), subset)
	assert.Subset(t, utils.GetMapKeys(body[1]), subset)

	afterEachGetPosts(ctx)
}

func TestGetPostsFailsIfUserIsNotLoggedIn(t *testing.T)  {
	ctx, cancel := beforeEachGetPosts()
	defer cancel()

	token = ""
	w := executeGetPosts()
	
	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	assert.Equal(t,http.StatusUnauthorized, w.Code)
	assert.Contains(t, body, "message")

	afterEachGetPosts(ctx)
}

func afterEachGetPosts(ctx context.Context)  {
	collection := app.GetCollectionHandle(handlers.PostCollection)
	collection.DeleteMany(ctx, bson.D{})
}

func executeGetPosts() *httptest.ResponseRecorder  {
	var router = routes.SetupRouter()
	body := map[string]string{"content": content, "category": category, "title": title}
	jsonString, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts", bytes.NewReader(jsonString))
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	router.ServeHTTP(w, req)

	return w
}

func beforeEachGetPosts() (context.Context, context.CancelFunc) {
	utils.LoadEnvVariables("../.env.test")
	ctx, cancel := app.CreateDataBaseConnection()

	user := models.User{ID: primitive.NewObjectID(), Email: "test@gmail.com"}
	token, _ = user.GenerateToken()
	
	post1 := models.Post{ID: primitive.NewObjectID(), Category: "Test", Title: "Test", Content: "Test", UserID: user.ID }
	post2 := models.Post{ID: primitive.NewObjectID(), Category: "Test", Title: "Test", Content: "Test", UserID: user.ID }
	post1.NormalizeFields(true)
	post2.NormalizeFields(true)
	
	collection := app.GetCollectionHandle(handlers.PostCollection)
	collection.InsertMany(ctx, []interface{}{post1, post2})
	return ctx, cancel
}