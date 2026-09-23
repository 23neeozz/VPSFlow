package errors

import (
	"errors"
	"fmt"
)

// Code represents a stable machine-readable error identifier.
type Code string

const (
	CodeInternal          Code = "internal_error"
	CodeValidation        Code = "validation_error"
	CodeNotFound          Code = "not_found"
	CodeConflict          Code = "conflict"
	CodeUnauthorized      Code = "unauthorized"
	CodeForbidden         Code = "forbidden"
	CodeRateLimited       Code = "rate_limited"
	CodeServiceUnavailable Code = "service_unavailable"
)

// AppError is the canonical application error type for VPSFlow services.
type AppError struct {
	Code    Code
	Message string
	Details map[string]any
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError.
func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap wraps an existing error with an AppError.
func Wrap(code Code, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

// WithDetails returns a copy of the error with additional details.
func (e *AppError) WithDetails(details map[string]any) *AppError {
	copy := *e
	if e.Details != nil {
		copy.Details = make(map[string]any, len(e.Details)+len(details))
		for k, v := range e.Details {
			copy.Details[k] = v
		}
	} else {
		copy.Details = make(map[string]any, len(details))
	}
	for k, v := range details {
		copy.Details[k] = v
	}
	return &copy
}

// IsCode reports whether err is an AppError with the given code.
func IsCode(err error, code Code) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// AsAppError attempts to extract an AppError from err.
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
