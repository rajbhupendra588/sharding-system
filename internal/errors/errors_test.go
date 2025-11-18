package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestError_Error(t *testing.T) {
	err := New(http.StatusNotFound, "resource not found")
	msg := err.Error()
	if msg != "resource not found" {
		t.Errorf("Expected 'resource not found', got %s", msg)
	}
}

func TestError_Error_WithWrappedError(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(originalErr, http.StatusInternalServerError, "wrapped error")
	msg := err.Error()
	
	if msg == "" {
		t.Error("Expected non-empty error message")
	}
	if !errors.Is(err, originalErr) {
		t.Error("Expected error to wrap original error")
	}
}

func TestError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(originalErr, http.StatusInternalServerError, "wrapped error")
	
	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original, got %v", unwrapped)
	}
}

func TestError_HTTPStatus(t *testing.T) {
	err := New(http.StatusNotFound, "not found")
	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, err.HTTPStatus())
	}
}

func TestNew(t *testing.T) {
	err := New(http.StatusBadRequest, "bad request")
	if err == nil {
		t.Fatal("Expected non-nil error")
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d, got %d", http.StatusBadRequest, err.Code)
	}
	if err.Message != "bad request" {
		t.Errorf("Expected message 'bad request', got %s", err.Message)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original")
	err := Wrap(originalErr, http.StatusInternalServerError, "wrapped")
	
	if err == nil {
		t.Fatal("Expected non-nil error")
	}
	if err.Err != originalErr {
		t.Errorf("Expected wrapped error to be original, got %v", err.Err)
	}
	if err.Code != http.StatusInternalServerError {
		t.Errorf("Expected code %d, got %d", http.StatusInternalServerError, err.Code)
	}
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected int
	}{
		{"NotFound", ErrNotFound, http.StatusNotFound},
		{"BadRequest", ErrBadRequest, http.StatusBadRequest},
		{"InternalServerError", ErrInternalServerError, http.StatusInternalServerError},
		{"Unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"Forbidden", ErrForbidden, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.HTTPStatus() != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, tt.err.HTTPStatus())
			}
		})
	}
}

