// Package validate provides functions for validating structs using the go-playground/validator library.
package validate

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// init initializes the validator instance.
func init() {
	validate = validator.New()
	validate.RegisterValidation("isUrl", urlValidator)
	validate.RegisterValidation("uniqueString", uniqueString)
	validate.RegisterValidation("uniqueStruct", uniqueStruct)
}

// Struct validates the given struct and returns a slice of errors, if any.
func Struct(s interface{}) []error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validateErrors := err.(validator.ValidationErrors)

	e := make([]error, len(validateErrors))

	for index, err := range validateErrors {
		field := strings.ToLower(err.StructField())
		switch err.Tag() {
		case "required":
			e[index] = fmt.Errorf("%s is required", field)
		case "min":
			e[index] = fmt.Errorf("%s must be greater than %s", field, err.Param())
		case "max":
			e[index] = fmt.Errorf("%s must be less than %s", field, err.Param())
		case "oneof":
			e[index] = fmt.Errorf("%s must be one of [%s]", field, err.Param())
		case "uuid4":
			e[index] = fmt.Errorf("%s must be a valid uuid4", field)
		default:
			e[index] = fmt.Errorf("%s %s", field, err.Tag())
		}
	}

	return e
}

func Var(field interface{}, tag string) error {
	err := validate.Var(field, tag)
	if err == nil {
		return nil
	}

	validateErrors := err.(validator.ValidationErrors)
	fmt.Println(validateErrors)
	return fmt.Errorf("%s", validateErrors.Error())
}

func Unmarshal(data []byte, s interface{}) error {
	var temp map[string]any
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if err := json.Unmarshal(data, s); err != nil {
		return err
	}

	return nil
}

func urlValidator(fl validator.FieldLevel) bool {
	u, err := url.Parse(fl.Field().String())
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func uniqueString(fl validator.FieldLevel) bool {
	values, ok := fl.Field().Interface().([]string)
	fmt.Println("VALUES: ", fl.Field())
	if !ok {
		return false
	}

	uniqueValues := make(map[string]bool)

	for _, v := range values {
		if uniqueValues[v] {
			return false
		}

		uniqueValues[v] = true
	}

	return true
}

func uniqueStruct(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() != reflect.Slice {
		return false
	}

	uniqueValues := make(map[string]bool)

	for i := 0; i < field.Len(); i++ {
		elem := field.Index(i).Interface()

		bytes, err := json.Marshal(elem)
		if err != nil {
			return false
		}

		stringJson := string(bytes)

		if _, exists := uniqueValues[stringJson]; exists {
			return false
		}

		uniqueValues[stringJson] = true
	}

	return true
}
