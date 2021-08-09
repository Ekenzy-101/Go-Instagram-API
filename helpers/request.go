package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

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
