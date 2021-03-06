// Overiding the gin framework's default validator which implements the StructValidator interface.
package helpers

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	usernameRegex = regexp.MustCompile("^[a-z0-9_][a-z0-9_.]{4,28}[a-z0-9_]$")
	nameRegex     = regexp.MustCompile("^[a-zA-Z][a-zA-z ]*$")
)

type DefaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

func (v *DefaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyinit()

		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}

	return nil
}

func (v *DefaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func (v *DefaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		err := v.validate.RegisterValidation("username", validateUserName)
		ExitIfError(err)

		err = v.validate.RegisterValidation("name", validateName)
		ExitIfError(err)

		err = v.validate.RegisterValidation("object_id", validateObjectID)
		ExitIfError(err)

		v.validate.RegisterTagNameFunc(jsonTagName)
	})
}

func jsonTagName(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func validateObjectID(fl validator.FieldLevel) bool {
	return primitive.IsValidObjectID(fl.Field().String())
}

func validateName(fl validator.FieldLevel) bool {
	return nameRegex.MatchString(fl.Field().String())
}

func validateUserName(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}
