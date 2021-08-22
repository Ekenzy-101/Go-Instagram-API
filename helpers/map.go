package helpers

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func GenerateErrorMessages(errors validator.ValidationErrors) map[string]string {
	messages := make(map[string]string)

	for _, err := range errors {
		field := err.Field()
		tag := err.ActualTag()
		switch tag {
		case "name":
			messages[field] = strings.Title(field) + " should contain only letters and space"
		case "email":
			messages[field] = strings.Title(field) + " is not a valid email address"
		case "gt":
			messages[field] = strings.Title(field) + " should be greater than " + err.Param()
		case "lte":
			messages[field] = strings.Title(field) + " should be less than or equal to " + err.Param()
		case "max":
			messages[field] = strings.Title(field) + " should be less than " + err.Param() + " characters"
		case "min":
			messages[field] = strings.Title(field) + " should be up to " + err.Param() + " characters"
		case "oneof":
			messages[field] = strings.Title(field) + " should be in these category " + err.Param()
		case "object_id":
			messages[field] = strings.Title(field) + " is not a valid ObjectID"
		case "required":
			messages[field] = strings.Title(field) + " is required"
		case "username":
			messages[field] = strings.Title(field) + " is not valid"
		default:
			messages[field] = strings.Title(field) + " is invalid"
		}
	}

	return messages
}
