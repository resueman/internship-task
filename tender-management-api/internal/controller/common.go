package controller

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	defaultLimit    = 5
	defaultOffset   = 0
	defaultUsername = ""
)

type errorResponse struct {
	Reason string `json:"reason"`
}

func getAllErrorMessages(err error) string {
	var builder strings.Builder
	for _, fe := range err.(validator.ValidationErrors) {
		message := fmt.Sprintf("'%s': %s\n", fe.Field(), getMessage(fe))
		builder.WriteString(message)
	}

	return builder.String()
}

func getMessage(fe validator.FieldError) string {
	s, i := "", int32(0)
	if fe.Type() == reflect.TypeOf(s) {
		return getMessageForString(fe)
	}

	if fe.Type() == reflect.TypeOf(i) {
		return getMessageForInt(fe)
	}

	if fe.Type() == reflect.TypeOf(0) {
		return getMessageForInt(fe)
	}

	return "Unknown error (2)"
}

func getMessageForInt(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "lte", "max":
		return "should be less or equal than " + fe.Param()
	case "gte", "min":
		return "should be greater or equal than " + fe.Param()
	}

	return "incorrect value passed"
}

func getMessageForString(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "lte", "max":
		return "length should be less or equal than " + fe.Param()
	case "gte", "min":
		return "length should be greater or equal than " + fe.Param()
	case "oneof":
		return "should have value in: " + fe.Param()
	}

	return "incorrect value passed"
}
