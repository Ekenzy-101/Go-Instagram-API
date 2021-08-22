package models

import (
	"time"

	"github.com/Ekenzy-101/Go-Gin-REST-API/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserDetails struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"  json:"_id"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	Followers       bson.A             `bson:"followers" json:"followers"`
	FollowersCount  int                `bson:"followersCount" json:"followersCount"`
	Following       bson.A             `bson:"following" json:"following"`
	FollowingCount  int                `bson:"followingCount" json:"followingCount"`
	SavedPosts      bson.A             `bson:"savedPosts" json:"savedPosts"`
	SavedPostsCount int                `bson:"savedPostsCount" json:"savedPostsCount"`
	UserID          interface{}        `bson:"userId,omitempty" json:"userId"`
}

func NewUserDetails(exclude []interface{}) bson.M {
	userDetails := bson.M{}

	if !helpers.Contains(exclude, "_id") {
		userDetails["_id"] = primitive.NewObjectID()
	}

	if !helpers.Contains(exclude, "createdAt") {
		userDetails["createdAt"] = time.Now()
	}

	if !helpers.Contains(exclude, "followers") {
		userDetails["followers"] = bson.A{}
	}

	if !helpers.Contains(exclude, "followersCount") {
		userDetails["followersCount"] = 0
	}

	if !helpers.Contains(exclude, "following") {
		userDetails["following"] = bson.A{}
	}

	if !helpers.Contains(exclude, "followingCount") {
		userDetails["followingCount"] = 0
	}
	if !helpers.Contains(exclude, "savedPosts") {
		userDetails["savedPosts"] = bson.A{}
	}

	if !helpers.Contains(exclude, "savedPostsCount") {
		userDetails["savedPostsCount"] = 0
	}

	return userDetails
}
