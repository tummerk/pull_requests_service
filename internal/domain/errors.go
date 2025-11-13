package domain

import (
	"fmt"
	"git.appkode.ru/pub/go/failure"
)

type AppError struct {
	Code    failure.ErrorCode
	Message string
	cause   error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.cause
}

func NewError(code failure.ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func WrapError(err error, code failure.ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		cause:   err,
	}
}
