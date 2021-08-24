package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Reply struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty" `
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	LikesCount int                `bson:"likesCount" json:"likesCount"`
	Message    string             `bson:"message" json:"message" binding:"required"`
	PostID     interface{}        `bson:"postId" json:"postId" binding:"object_id"`
	ReplyToID  interface{}        `bson:"replyToId" json:"replyToId" binding:"object_id"`
	User       bson.M             `bson:"user,omitempty" json:"user,omitempty"`
	UserID     interface{}        `bson:"userId,omitempty" json:"userId,omitempty"`
}

func (reply *Reply) NormalizeFields(userId primitive.ObjectID) error {
	reply.ID = primitive.NewObjectID()
	reply.CreatedAt = time.Now()
	reply.UserID = userId

	var err error
	reply.PostID, err = primitive.ObjectIDFromHex(fmt.Sprintf("%v", reply.PostID))
	if err != nil {
		return err
	}

	reply.ReplyToID, err = primitive.ObjectIDFromHex(fmt.Sprintf("%v", reply.ReplyToID))
	if err != nil {
		return err
	}

	return nil
}

func (reply *Reply) SetUser(user *User) {
	reply.UserID = nil
	reply.User = bson.M{
		"username": user.Username,
		"image":    user.Image,
	}
}
