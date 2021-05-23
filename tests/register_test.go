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

var (
	email string
	password string
	name string
)
func TestRegisterSucceedsWithValidInputs(t *testing.T)  {
	ctx, cancel := beforeEachRegister()
	defer cancel()

	w := executeRegister()
	
	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	subset := []string{"_id", "name", "email"}
	assert.Equal(t,http.StatusCreated, w.Code)
	assert.Subset(t, utils.GetMapKeys(body), subset)
	assert.Contains(t, w.Result().Header, "Set-Cookie")

	afterEachRegister(ctx)
}
func TestRegisterFailsWithInvalidInputs(t *testing.T)  {
	ctx, cancel := beforeEachRegister()
	defer cancel()

	email = strings.Join(make([]string, 247), "a") + "@gmail.com"
	password = "111"
	name = ""
	w := executeRegister()
	
	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	subset := []string{"password", "name", "email"}
	assert.Equal(t,http.StatusBadRequest, w.Code)
	assert.Subset(t, utils.GetMapKeys(body), subset)

	afterEachRegister(ctx)
}
func TestRegisterFailsIfUserExistsInDatabase(t *testing.T)  {
	ctx, cancel := beforeEachRegister()
	defer cancel()

	user := models.User{
		Name : name,
		Email: email,
		Password: password,	
	}
	collection := app.GetCollectionHandle(handlers.UserCollection)
	collection.InsertOne(ctx, &user)
	
	w := executeRegister()
	
	body := map[string]string{}
	json.NewDecoder(w.Body).Decode(&body)
	
	assert.Equal(t,http.StatusBadRequest, w.Code)
	assert.Contains(t, body, "message")

	afterEachRegister(ctx)
}

func afterEachRegister(ctx context.Context)  {
	collection := app.GetCollectionHandle(handlers.UserCollection)
	collection.DeleteMany(ctx, bson.D{})
}

func executeRegister() *httptest.ResponseRecorder  {
	var router = routes.SetupRouter()
	body := map[string]string{"email": email, "password": password, "name": name}
	jsonString, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonString))
	router.ServeHTTP(w, req)

	return w
}

func beforeEachRegister() (context.Context, context.CancelFunc) {
	utils.LoadEnvVariables("../.env.test")
	ctx, cancel := app.CreateDataBaseConnection()
	email = "test@gmail.com"
	password = "123456"
	name = "test"
	
	return ctx, cancel
}
