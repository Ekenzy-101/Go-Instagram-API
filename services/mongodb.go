package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

var mongoClient *mongo.Client

func CreateMongoDBConnection() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoDBURI))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	mongoClient = client
	if !config.IsTesting {
		fmt.Println("Successfully connected and pinged")
	}

	createIndexes(ctx)
}

func CloseMongoDBConnection() error {
	return mongoClient.Disconnect(context.TODO())
}

func createIndexes(ctx context.Context) ([]string, error) {
	usersCollection := GetMongoDBCollection(config.UsersCollection)
	postsCollection := GetMongoDBCollection(config.PostsCollection)

	userModels := []mongo.IndexModel{{
		Keys:    bsonx.Doc{{Key: "email", Value: bsonx.Int32(1)}},
		Options: options.Index().SetUnique(true),
	},
		{
			Keys:    bsonx.Doc{{Key: "username", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		}}
	userIndexes, err := usersCollection.Indexes().CreateMany(ctx, userModels)
	if err != nil {
		return nil, err
	}

	postModels := []mongo.IndexModel{{
		Keys:    bsonx.Doc{{Key: "createdAt", Value: bsonx.Int32(1)}},
		Options: options.Index(),
	}}
	postIndexes, err := postsCollection.Indexes().CreateMany(ctx, postModels)
	if err != nil {
		return nil, err
	}

	indexes := append(userIndexes, postIndexes...)
	return indexes, nil
}

func GetMongoDBCollection(name string) *mongo.Collection {
	return mongoClient.Database(config.MongoDBName).Collection(name)
}
