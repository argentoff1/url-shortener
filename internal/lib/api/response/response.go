package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOk    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOk,
	}
}

func Error(message string) Response {
	return Response{
		Status: StatusError,
		Error:  message,
	}
}

// ValidationError - формирование запроса в более понятные ошибки для пользователя
func ValidationError(errs validator.ValidationErrors) Response {
	var errMessages []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMessages = append(errMessages, fmt.Sprintf(
				"Поле %s обязательное для заполнения", err.Field()))
		case "url":
			errMessages = append(errMessages, fmt.Sprintf(
				"Поле %s не является допустимым URL адресом", err.Field()))
		default:
			errMessages = append(errMessages, fmt.Sprintf(
				"Поле %s не является допустимым", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMessages, ", "),
	}
}
