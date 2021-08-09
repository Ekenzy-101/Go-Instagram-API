package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	AccessTokenTTLInSeconds = 3600
	PostsCollection         = "posts"
	UsersCollection         = "users"
)

var (
	AccessTokenSecret string
	ClientOrigin      string
	MongoDBURI        string
	MongoDBName       string
	Port              string
	IsDevelopment     = gin.Mode() == gin.DebugMode
	IsProduction      = gin.Mode() == gin.ReleaseMode
	IsTesting         = gin.Mode() == gin.TestMode
)

func init() {
	filename := ""
	if IsTesting {
		filename = "../.env.test"
	}

	if IsDevelopment {
		filename = ".env"
	}

	if filename != "" {
		if err := godotenv.Load(filename); err != nil {
			log.Println(err)
		}
	}

	AccessTokenSecret = os.Getenv("ACCESS_TOKEN_SECRET")
	ClientOrigin = os.Getenv("CLIENT_ORIGIN")
	MongoDBName = os.Getenv("MONGODB_NAME")
	MongoDBURI = os.Getenv("MONGODB_URI")
	Port = os.Getenv("PORT")
	if Port == "" {
		Port = "5000"
	}
}
