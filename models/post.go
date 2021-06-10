package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty" `
	Category  string             `bson:"category" json:"category,omitempty"  binding:"required"`
	Content   string             `bson:"content" json:"content,omitempty"  binding:"required"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt,omitempty"`
	Title     string             `bson:"title" json:"title,omitempty"  binding:"required"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt,omitempty"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId,omitempty"`
}

func (post *Post) NormalizeFields(withTimestamps bool) {
	post.Category = strings.TrimSpace(post.Category)
	post.Content = strings.TrimSpace(post.Content)
	post.Title = strings.TrimSpace(post.Title)

	if withTimestamps {
		post.CreatedAt = time.Now()
		post.UpdatedAt = time.Now()
	}
}

func (post *Post) UpdateFields(body *Post) {
	post.Category = body.Category
	post.Content = body.Content
	post.Title = body.Title
}
