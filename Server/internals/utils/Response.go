package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response[T any] struct {
	Status  int    `json:"status"`
	Data    T      `json:"data,omitempty"`
	Message string `json:"message"`
	Heading string `json:"heading"`
}

func WriteJson[T any](response http.ResponseWriter, status int, data Response[T]) error {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)

	return json.NewEncoder(response).Encode(data)
}

func GeneralError(err error, message string, status int, heading string) Response[string] {
	return Response[string]{
		Data:    err.Error(),
		Status:  status,
		Message: message,
		Heading: heading,
	}
}

func ValidationError(errs validator.ValidationErrors) Response[string] {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is  invalid", err.Field()))
		}
	}

	return Response[string]{
		Status:  422,
		Data:    strings.Join(errMsgs, ", "),
		Heading: "Please provide all the required data",
		Message: strings.Join(errMsgs, ", "),
	}

}
