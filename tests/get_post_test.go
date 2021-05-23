package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	postId string
)

func TestGetPostSucceedsIfUserIsLoggedIn(t *testing.T)  {
	ctx, cancel := beforeEachGetPost()
	defer cancel()

	w := executeGetPost()
	
	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	subset := []string{"_id", "content", "category", "title", "userId", "createdAt", "updatedAt"}
	assert.Equal(t,http.StatusOK, w.Code)
	assert.Subset(t, utils.GetMapKeys(body), subset)

	afterEachGetPost(ctx)
}

func TestGetPostFailsIfUserIsNotLoggedIn(t *testing.T)  {
	ctx, cancel := beforeEachGetPost()
	defer cancel()

	token = ""
	w := executeGetPost()
	
	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	assert.Equal(t,http.StatusUnauthorized, w.Code)
	assert.Contains(t, body, "message")

	afterEachGetPost(ctx)
}

func TestGetPostFailsIfPostIdIsInvalid(t *testing.T)  {
	ctx, cancel := beforeEachGetPost()
	defer cancel()

	postId = "invalid"
	w := executeGetPost()
	
	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	assert.Equal(t,http.StatusBadRequest, w.Code)
	assert.Contains(t, strings.ToLower(body["message"]), "invalid")

	afterEachGetPost(ctx)
}

func TestGetPostFailsIfPostNotFound(t *testing.T)  {
	ctx, cancel := beforeEachGetPost()
	defer cancel()

	postId = primitive.NewObjectID().Hex()
	w := executeGetPost()
	
	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	assert.Equal(t,http.StatusNotFound, w.Code)
	assert.Contains(t, strings.ToLower(body["message"]), "not found")

	afterEachGetPost(ctx)
}

func afterEachGetPost(ctx context.Context)  {
	collection := app.GetCollectionHandle(handlers.PostCollection)
	collection.DeleteMany(ctx, bson.D{})
}

func executeGetPost() *httptest.ResponseRecorder  {
	var router = routes.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/" + postId, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	router.ServeHTTP(w, req)

	return w
}

func beforeEachGetPost() (context.Context, context.CancelFunc) {
	utils.LoadEnvVariables("../.env.test")
	ctx, cancel := app.CreateDataBaseConnection()

	user := &models.User{ID: primitive.NewObjectID(), Email: "test@gmail.com"}
	token, _ = user.GenerateToken()
	
	post := &models.Post{ID: primitive.NewObjectID(), Category: "Test", Title: "Test", Content: "Test", UserID: user.ID }
	post.NormalizeFields(true)
	postId = post.ID.Hex()

	collection := app.GetCollectionHandle(handlers.PostCollection)
	collection.InsertOne(ctx, post)
	return ctx, cancel
}