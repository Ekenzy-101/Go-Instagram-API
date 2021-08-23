package models

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/config"
	"github.com/Ekenzy-101/Go-Gin-REST-API/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	PostProjection = bson.M{"images": 1, "likesCount": 1, "commentsCount": 1, "createdAt": 1}
)

type Post struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty" `
	Caption       string             `bson:"caption" json:"caption"`
	Comments      []Comment          `bson:"comments" json:"comments"`
	CommentsCount int                `bson:"commentsCount" json:"commentsCount"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	Images        []string           `bson:"images" json:"images"`
	ImageCount    int                `bson:"imageCount,omitempty" json:"imageCount,omitempty" binding:"gt=0"`
	LikesCount    int                `bson:"likesCount" json:"likesCount"`
	Location      string             `bson:"location" json:"location"`
	User          bson.M             `bson:"user,omitempty" json:"user"`
	UserID        interface{}        `bson:"userId,omitempty" json:"userId,omitempty"`
}

func (post *Post) GeneratePresignedURLKeys() []string {
	keys := make([]string, post.ImageCount)

	for index := range keys {
		keys[index] = post.ID.Hex() + "/" + strconv.Itoa(index)
	}

	return keys
}

func (post *Post) GetCommentIds() bson.A {
	commentIds := bson.A{}
	for _, comment := range post.Comments {
		commentIds = append(commentIds, comment.ID)
	}

	return commentIds
}

func (post *Post) NormalizeFields(userId interface{}) {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.UserID = userId

	if post.Comments == nil {
		post.Comments = []Comment{}
	}
}

func (post *Post) SetUser(user *User) {
	post.UserID = nil
	post.User = bson.M{
		"username": user.Username,
		"image":    user.Image,
	}
}

func (post *Post) SetImages(urls []string) {
	post.ImageCount = 0
	post.Images = make([]string, len(urls))

	for index, url := range urls {
		post.Images[index] = strings.Split(url, "?")[0]
	}
}

type FindPostResult struct {
	Post         *Post
	ResponseBody interface{}
	StatusCode   int
}

func FindPost(ctx context.Context, filter interface{}, options ...*options.FindOneOptions) *FindPostResult {
	post := &Post{}
	collection := services.GetMongoDBCollection(config.PostsCollection)
	err := collection.FindOne(ctx, filter, options...).Decode(post)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return &FindPostResult{
			ResponseBody: gin.H{"message": "Post not found"},
			StatusCode:   http.StatusNotFound,
		}
	}

	if err != nil {
		return &FindPostResult{
			ResponseBody: gin.H{"message": err.Error()},
			StatusCode:   http.StatusInternalServerError,
		}
	}

	return &FindPostResult{
		Post: post,
	}
}

func MapPostsToUserSubDocuments(posts ...Post) bson.A {
	postDocuments := make(bson.A, len(posts))

	for index, post := range posts {
		postDocument := bson.M{
			"_id":           post.ID,
			"images":        post.Images,
			"likesCount":    post.LikesCount,
			"commentsCount": post.CommentsCount,
			"createdAt":     post.CreatedAt,
		}
		postDocuments[index] = postDocument
	}

	return postDocuments
}
