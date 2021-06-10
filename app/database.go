package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

var mongoClient *mongo.Client

func CreateDataBaseConnection() (context.Context, context.CancelFunc) {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}

	mongoClient = client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connected to MongoDB!")
	}

	CreateIndexes(ctx)

	return ctx, cancel
}

func CreateIndexes(ctx context.Context) (string, error) {
	collection := mongoClient.Database(os.Getenv("MONGODB_NAME")).Collection("users")
	return collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bsonx.Doc{{Key :"email", Value: bsonx.Int32(1)}},
		Options: options.Index().SetUnique(true),
	})
}

func GetCollectionHandle(name string) *mongo.Collection {
	return mongoClient.Database(os.Getenv("MONGODB_NAME")).Collection(name)
}
