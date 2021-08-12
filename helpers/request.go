package helpers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoErrorOption struct {
	Code    int
	Message string
	Err     error
}

func ValidateFindOneError(option MongoErrorOption) (code int, responseBody interface{}) {
	if errors.Is(option.Err, mongo.ErrNoDocuments) {
		return option.Code, gin.H{"message": option.Message}
	}

	return 200, nil
}

func ValidateMongoError(option MongoErrorOption) (code int, responseBody interface{}) {
	// Microservice to log error
	if mongo.IsTimeout(option.Err) {
		return http.StatusInternalServerError, gin.H{"message": "We couldn't complete your request. Please try again", "error": "Timeout Error"}
	}

	if mongo.IsNetworkError(option.Err) {
		return http.StatusInternalServerError, gin.H{"message": "We could", "error": "Network Error"}
	}

	return 200, nil
}

func ValidateRequestBody(c *gin.Context, body interface{}) interface{} {
	// Body should be a pointer to a value
	err := c.ShouldBindJSON(body)
	validationErrors := validator.ValidationErrors{}
	if errors.As(err, &validationErrors) {
		return GenerateErrorMessages(validationErrors)
	}

	if err != nil {
		return gin.H{"message": err.Error()}
	}

	return nil
}
