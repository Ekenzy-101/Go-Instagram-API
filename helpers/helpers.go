package helpers

import (
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func GenerateErrorMessages(errors validator.ValidationErrors) map[string]string {
	messages := make(map[string]string)

	for _, err := range errors {
		field := err.Field()
		tag := err.ActualTag()
		switch tag {
		case "alpha":
			messages[field] = strings.Title(field) + " should contain only letters"
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
		case "required":
			messages[field] = strings.Title(field) + " is required"
		case "username":
			messages[field] = strings.Title(field) + " is not valid"
		default:
			messages[field] = "Field is invalid"
		}
	}

	return messages
}

func GetMapKeys(value map[string]string) []string {
	keys := make([]string, len(value))

	i := 0
	for key := range value {
		keys[i] = key
		i++
	}

	return keys
}

func LoadEnvVariables(filenames ...string) {
	if err := godotenv.Load(filenames...); err != nil {
		log.Fatal(err)
	}
}
