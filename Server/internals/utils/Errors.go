package utils

import "errors"

type ErrorType int

const (
	NoError ErrorType = iota
	ErrConflict
	ErrInternal
	ErrUnauthorized
	ErrNotFound
	ErrBadRequest
	ErrForBidden
	ErrUnprocessableData
)

type ErrorValue struct {
	Message    string
	StatusCode int
}

var ErrorMap = map[ErrorType]ErrorValue{
	NoError:              {Message: "No Error", StatusCode: 200},
	ErrConflict:          {Message: "Resource Conflict Try again later ... ", StatusCode: 409},
	ErrInternal:          {Message: "Internal Server Error ...", StatusCode: 500},
	ErrUnauthorized:      {Message: "Unauthorized ", StatusCode: 401},
	ErrNotFound:          {Message: "Not Found", StatusCode: 404},
	ErrBadRequest:        {Message: "BadRequest", StatusCode: 400},
	ErrForBidden:         {Message: "You are Forbidden for this service", StatusCode: 403},
	ErrUnprocessableData: {Message: "Please provide a valid/processable data", StatusCode: 422},
}

func GenerateError(errType ErrorType, err error) Response[string] {
	if _, ok := ErrorMap[errType]; !ok {
		return GeneralError(
			errors.New("Invalid Err Type recieved"), "Invalid err type ", 500, ErrorMap[ErrInternal].Message)
	}
	return GeneralError(err, err.Error(), ErrorMap[errType].StatusCode, ErrorMap[errType].Message)
}
