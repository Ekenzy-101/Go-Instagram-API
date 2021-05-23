package models

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/argon2"
)

type JwtClaim struct {
	Email string `json:"email"`
	ID string `json:"_id"`
	jwt.StandardClaims
}

type PasswordConfig struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

type User struct {
	ID          primitive.ObjectID	`bson:"_id,omitempty"  json:"_id,omitempty"`
	CreatedAt   time.Time          	`bson:"createdAt" json:"createdAt,omitempty"`
	Email       string             	`bson:"email" json:"email,omitempty" binding:"email,max=255"`
	Name        string             	`bson:"name" json:"name,omitempty" binding:"required,max=50"`
	Password	  string             	`bson:"password" json:"password,omitempty"  binding:"required,min=6"`
	UpdatedAt   time.Time          	`bson:"updatedAt" json:"updatedAt,omitempty"`
}

func (user *User) ComparePassword(password string) (bool, error) {
	parts := strings.Split(user.Password, "$")

	if len(parts) < 4 {
		return false, errors.New("invalid string")
	}

	c := &PasswordConfig{}
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &c.memory, &c.time, &c.threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	c.keyLen = uint32(len(decodedHash))

	comparisonHash := argon2.IDKey([]byte(password), salt, c.time, c.memory, c.threads, c.keyLen)

	return (subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1), nil
}

func (user *User) GenerateToken() (string, error) {
	claims := &JwtClaim{
		Email: user.Email,
		ID : user.ID.Hex(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(1)).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(os.Getenv("APP_ACCESS_SECRET")))
	return signedToken, err
}

func (user *User) HashPassword() error {
	c := &PasswordConfig{
		time:    1,
		memory:  64 * 1024,
		threads: 4,
		keyLen:  32,
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	hash := argon2.IDKey([]byte(user.Password), salt, c.time, c.memory, c.threads, c.keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	format := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	user.Password = fmt.Sprintf(format, argon2.Version, c.memory, c.time, c.threads, b64Salt, b64Hash)

	return nil
}

func (user *User) NormalizeFields(withTimestamps bool) {
	user.Email = strings.ToLower(user.Email)
	user.Name = strings.TrimSpace(user.Name)

	if withTimestamps {
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	}
}

func VerifyToken(signedToken string) (claims *JwtClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("APP_ACCESS_SECRET")), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*JwtClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("JWT is expired")
		return
	}

	return
}