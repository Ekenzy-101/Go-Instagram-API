package tests

import (
	"bytes"
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
)

func TestLoginSucceedsWithValidInputs(t *testing.T) {
	ctx, cancel := beforeEachLogin()
	defer cancel()

	w := executeLogin()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	subset := []string{"_id", "name", "email"}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Subset(t, utils.GetMapKeys(body), subset)
	assert.Contains(t, w.Result().Header, "Set-Cookie")

	afterEachLogin(ctx)
}
func TestLoginFailsWithInvalidInputs(t *testing.T) {
	ctx, cancel := beforeEachLogin()
	defer cancel()

	email = strings.Join(make([]string, 247), "a") + "@gmail.com"
	password = "111"
	w := executeLogin()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	subset := []string{"email", "password"}
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Subset(t, utils.GetMapKeys(body), subset)

	afterEachLogin(ctx)
}
func TestLoginFailsIfUserDoesNotExistInDatabase(t *testing.T) {
	ctx, cancel := beforeEachLogin()
	defer cancel()

	email = "doesnotexist@gmail.com"
	w := executeLogin()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, body, "message")

	afterEachLogin(ctx)
}
func TestLoginFailsIfPasswordDoesNotMatch(t *testing.T) {
	ctx, cancel := beforeEachLogin()
	defer cancel()

	password = "notmatch"
	w := executeLogin()

	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, body, "message")

	afterEachLogin(ctx)
}

func afterEachLogin(ctx context.Context) {
	collection := app.GetCollectionHandle(handlers.UserCollection)
	collection.DeleteMany(ctx, bson.D{})
}

func executeLogin() *httptest.ResponseRecorder {
	var router = routes.SetupRouter()
	body := map[string]string{"email": email, "password": password}
	jsonString, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(jsonString))
	router.ServeHTTP(w, req)

	return w
}

func beforeEachLogin() (context.Context, context.CancelFunc) {
	utils.LoadEnvVariables("../.env.test")
	ctx, cancel := app.CreateDataBaseConnection()
	email = "test@gmail.com"
	password = "123456"
	user := models.User{
		Name:     "test",
		Email:    email,
		Password: password,
	}
	user.HashPassword()

	collection := app.GetCollectionHandle(handlers.UserCollection)
	collection.InsertOne(ctx, &user)

	return ctx, cancel
}
