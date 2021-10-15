package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"go.mongodb.org/mongo-driver/bson"
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
	helpers.ExitIfError(err)

	helpers.ExitIfError(client.Ping(ctx, readpref.Primary()))

	mongoClient = client
	if !config.IsTesting {
		fmt.Println("Successfully connected and pinged")
	}

	_, err = CreateMongoDBIndexes(ctx)
	helpers.ExitIfError(err)

	if config.IsTesting {
		DeleteAllMongoDBDocuments()
	}
}

func CloseMongoDBConnection() error {
	return mongoClient.Disconnect(context.TODO())
}

func CreateMongoDBIndexes(ctx context.Context) ([]string, error) {
	userModels := []mongo.IndexModel{{
		Keys:    bsonx.Doc{{Key: "email", Value: bsonx.Int32(1)}},
		Options: options.Index().SetUnique(true),
	},
		{
			Keys:    bsonx.Doc{{Key: "username", Value: bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		}}
	usersCollection := GetMongoDBCollection(config.UsersCollection)
	userIndexes, err := usersCollection.Indexes().CreateMany(ctx, userModels)
	if err != nil {
		return nil, err
	}

	postModels := []mongo.IndexModel{{
		Keys: bsonx.Doc{{Key: "userId", Value: bsonx.Int32(1)}, {Key: "createdAt", Value: bsonx.Int32(-1)}},
	}, {
		Keys: bsonx.Doc{{Key: "caption", Value: bsonx.String("text")}, {Key: "createdAt", Value: bsonx.Int32(-1)}},
	}}
	postsCollection := GetMongoDBCollection(config.PostsCollection)
	postIndexes, err := postsCollection.Indexes().CreateMany(ctx, postModels)
	if err != nil {
		return nil, err
	}

	userDetailModels := []mongo.IndexModel{{
		Keys: bsonx.Doc{{Key: "userId", Value: bsonx.Int32(1)}, {Key: "createdAt", Value: bsonx.Int32(-1)}},
	}}
	userDetailsCollection := GetMongoDBCollection(config.UserDetailsCollection)
	userDetailIndexes, err := userDetailsCollection.Indexes().CreateMany(ctx, userDetailModels)
	if err != nil {
		return nil, err
	}

	commentModels := []mongo.IndexModel{{
		Keys: bsonx.Doc{{Key: "postId", Value: bsonx.Int32(1)}, {Key: "createdAt", Value: bsonx.Int32(-1)}},
	}}
	commentsCollection := GetMongoDBCollection(config.CommentsCollection)
	commentIndexes, err := commentsCollection.Indexes().CreateMany(ctx, commentModels)
	if err != nil {
		return nil, err
	}

	replyModels := []mongo.IndexModel{{
		Keys: bsonx.Doc{{Key: "postId", Value: bsonx.Int32(1)}},
	}, {
		Keys: bsonx.Doc{{Key: "replyToId", Value: bsonx.Int32(1)}, {Key: "createdAt", Value: bsonx.Int32(-1)}},
	}}
	repliesCollection := GetMongoDBCollection(config.RepliesCollection)
	replyIndexes, err := repliesCollection.Indexes().CreateMany(ctx, replyModels)
	if err != nil {
		return nil, err
	}

	indexes := append(userIndexes, postIndexes...)
	indexes = append(indexes, userDetailIndexes...)
	indexes = append(indexes, commentIndexes...)
	indexes = append(indexes, replyIndexes...)
	return indexes, nil
}

func GetMongoDBSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return mongoClient.StartSession(opts...)
}

func GetMongoDBClient() *mongo.Client {
	return mongoClient
}

func GetMongoDBCollection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return mongoClient.Database(config.MongoDBName).Collection(name, opts...)
}

func DeleteAllMongoDBDocuments() {
	commentsCollection := GetMongoDBCollection(config.CommentsCollection)
	_, err := commentsCollection.DeleteMany(context.Background(), bson.M{})
	helpers.ExitIfError(err)

	postsCollection := GetMongoDBCollection(config.PostsCollection)
	_, err = postsCollection.DeleteMany(context.Background(), bson.M{})
	helpers.ExitIfError(err)

	repliesCollection := GetMongoDBCollection(config.RepliesCollection)
	_, err = repliesCollection.DeleteMany(context.Background(), bson.M{})
	helpers.ExitIfError(err)

	usersCollection := GetMongoDBCollection(config.UsersCollection)
	_, err = usersCollection.DeleteMany(context.Background(), bson.M{})
	helpers.ExitIfError(err)

	userDetailsCollection := GetMongoDBCollection(config.UserDetailsCollection)
	_, err = userDetailsCollection.DeleteMany(context.Background(), bson.M{})
	helpers.ExitIfError(err)

	time.Sleep(1 * time.Second)
}
