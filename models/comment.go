package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Comment struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty" `
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	LikesCount   int                `bson:"likesCount" json:"likesCount"`
	Message      string             `bson:"message" json:"message" binding:"required"`
	PostID       interface{}        `bson:"postId" json:"postId" binding:"object_id"`
	RepliesCount int                `bson:"repliesCount" json:"repliesCount"`
	User         bson.M             `bson:"user,omitempty" json:"user,omitempty"`
	UserID       interface{}        `bson:"userId,omitempty" json:"userId,omitempty"`
}

func (comment *Comment) NormalizeFields(userId primitive.ObjectID) error {
	var err error
	comment.CreatedAt = time.Now()
	comment.ID = primitive.NewObjectID()
	comment.UserID = userId
	comment.PostID, err = primitive.ObjectIDFromHex(fmt.Sprintf("%v", comment.PostID))
	if err != nil {
		return err
	}
	return nil
}

func (comment *Comment) SetUser(user *User) {
	comment.UserID = nil
	comment.User = bson.M{
		"username": user.Username,
		"image":    user.Image,
	}
}

type FindCommentResult struct {
	Error        error
	Comment      *Comment
	ResponseBody interface{}
	StatusCode   int
}

func FindComment(ctx context.Context, filter interface{}, options ...*options.FindOneOptions) *FindCommentResult {
	comment := &Comment{}
	collection := services.GetMongoDBCollection(config.CommentsCollection)
	err := collection.FindOne(ctx, filter, options...).Decode(comment)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return &FindCommentResult{
			ResponseBody: gin.H{"message": "Comment not found"},
			StatusCode:   http.StatusNotFound,
			Error:        err,
		}
	}

	if err != nil {
		return &FindCommentResult{
			ResponseBody: gin.H{"message": err.Error()},
			StatusCode:   http.StatusInternalServerError,
			Error:        err,
		}
	}

	return &FindCommentResult{
		Comment: comment,
	}
}
