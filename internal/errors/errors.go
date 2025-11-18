package errors

import (
	"fmt"
	"net/http"
)

// Error represents an application error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new error
func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error
func Wrap(err error, code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error constructors
var (
	ErrNotFound            = New(http.StatusNotFound, "resource not found")
	ErrBadRequest          = New(http.StatusBadRequest, "bad request")
	ErrInternalServerError = New(http.StatusInternalServerError, "internal server error")
	ErrUnauthorized        = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden           = New(http.StatusForbidden, "forbidden")
)

// HTTPStatus returns the HTTP status code for the error
func (e *Error) HTTPStatus() int {
	return e.Code
}

