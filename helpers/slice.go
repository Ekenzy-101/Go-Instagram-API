package helpers

import (
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

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

func GetStructFields(value interface{}, exclude bson.A) bson.A {
	reflectionType := reflect.TypeOf(value)

	fields := bson.A{}
	for i := 0; i < reflectionType.NumField(); i++ {
		field := strings.Split(reflectionType.Field(i).Tag.Get("json"), ",")[0]
		if !Contains(exclude, field) {
			fields = append(fields, field)
		}
	}

	return fields
}
