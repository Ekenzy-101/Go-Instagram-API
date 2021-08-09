package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	email          = "test@gmail.com"
	hashedPassword string
	password       = "password"
	signedToken    string
	userId         = primitive.NewObjectID()
)

func TestHashPassword(t *testing.T) {
	user := &User{
		Password: password,
	}
	err := user.HashPassword()
	hashedPassword = user.Password

	assert.NoError(t, err)
}
func TestComparePassword(t *testing.T) {
	user := &User{
		Password: hashedPassword,
	}
	ok, err := user.ComparePassword(password)

	assert.NoError(t, err)
	assert.True(t, ok)
}
func TestGenerateToken(t *testing.T) {
	user := &User{
		Email: email,
		ID:    userId,
	}
	token, err := user.GenerateToken()
	signedToken = token

	assert.NoError(t, err)
}
