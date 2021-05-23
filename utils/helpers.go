package utils

import (
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func GenerateErrorMessages(errors validator.ValidationErrors) map[string]string  {
	messages := make(map[string]string)

	for _, err := range errors {
		field := strings.ToLower(err.Field())
		tag := err.ActualTag()
		switch  {
			case strings.Contains(tag, "required") :
				messages[field] = err.Field() + " is required"
			case strings.Contains(tag, "email") :
				messages[field] = err.Field() + " is not a valid email address"
			case strings.Contains(tag, "max") :
				messages[field] = err.Field() + " should be less than " + err.Param() + " characters"
			case strings.Contains(tag, "min") :
				messages[field] = err.Field() + " should be up to " + err.Param() + " characters"
			default:
				messages[field] = "Field is invalid"
		}
	}

	return messages
}


func GetMapKeys(value map[string]string) []string {
	log.Println(len(value))
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
