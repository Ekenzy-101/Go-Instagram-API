package models

import (
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty" `
	Images        []string           `bson:"images" json:"images"`
	Caption       string             `bson:"caption" json:"caption"`
	Location      string             `bson:"location" json:"location"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UserID        interface{}        `bson:"userId,omitempty" json:"userId,omitempty"`
	User          bson.M             `bson:"user,omitempty" json:"user"`
	CommentsCount int                `bson:"commentsCount" json:"commentsCount"`
	LikesCount    int                `bson:"likesCount" json:"likesCount"`
	ImageCount    int                `bson:"imageCount,omitempty" json:"imageCount,omitempty" binding:"gt=0"`
}

func (post *Post) GeneratePresignedURLKeys() []string {
	keys := make([]string, post.ImageCount)

	for index := range keys {
		keys[index] = post.ID.Hex() + "/" + strconv.Itoa(index)
	}

	return keys
}

func (post *Post) NormalizeFields(user *User) {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.UserID = &user.ID
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
