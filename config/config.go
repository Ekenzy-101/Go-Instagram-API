package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	AccessTokenCookieName   = "access_token"
	AccessTokenTTLInSeconds = 3600
	CommentsCollection      = "comments"
	RepliesCollection       = "replies"
	CommonPaginationLength  = 12
	LargePaginationLength   = 2 // TODO: Change later to 1200
	UserDetailsCollection   = "user_details"
	PostsCollection         = "posts"
	UsersCollection         = "users"
)

var (
	AWSBucket         string
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
		filename = "../../.env.test"
	}

	if IsDevelopment {
		filename = ".env"
	}

	if filename != "" {
		if err := godotenv.Load(filename); err != nil {
			log.Println(err)
		}
	}

	AWSBucket = os.Getenv("AWS_BUCKET")
	AccessTokenSecret = os.Getenv("ACCESS_TOKEN_SECRET")
	ClientOrigin = os.Getenv("CLIENT_ORIGIN")
	MongoDBName = os.Getenv("MONGODB_NAME")
	MongoDBURI = os.Getenv("MONGODB_URI")
	Port = os.Getenv("PORT")
	if Port == "" {
		Port = "5000"
	}
}
