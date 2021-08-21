package helpers

import (
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ExitIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

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
			messages[field] = "Field is invalid"
		}
	}

	return messages
}

func GetMapKeys(object interface{}) []string {
	reflectValue := reflect.ValueOf(object)
	if reflectValue.Kind() != reflect.Map {
		panic("value must be a map")
	}

	reflectType := reflect.TypeOf(object)
	if reflectType.Key().Kind() != reflect.String {
		panic("key must be a string")
	}

	keys := []string{}
	for _, key := range reflectValue.MapKeys() {
		keys = append(keys, key.String())
	}

	return keys
}

func GetMapValues(object interface{}) []interface{} {
	reflectValue := reflect.ValueOf(object)
	if reflectValue.Kind() != reflect.Map {
		panic("value must be a map")
	}

	iter := reflectValue.MapRange()
	values := []interface{}{}
	for iter.Next() {
		values = append(values, iter.Value().Interface())
	}

	return values
}
