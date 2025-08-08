package errors

import (
	"context"
	"fmt"
	"runtime"

	"github.com/diabolusgx/guess-the-number/pkg/logger"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	ErrCodeMissingPermissions ErrorCode = "missing_permissions"
	ErrCodeValidation         ErrorCode = "validation_error"
	ErrCodeNotFound           ErrorCode = "not_found"
	ErrCodeAlreadyExists      ErrorCode = "already_exists"
	ErrCodePermissionDenied   ErrorCode = "permission_denied"
	ErrCodeDatabase           ErrorCode = "database_error"
	ErrCodeInternalError      ErrorCode = "internal_error"
)

func (e ErrorCode) String() string {
	return string(e)
}

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode
	Message    string
	Details    map[string]any
	InnerError error
	Context    context.Context
}

func (e *AppError) Error() string {
	if e.InnerError != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.InnerError)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: make(map[string]any),
	}
}

// WithError wraps an existing error
func WithError(err error) *AppError {
	return &AppError{
		InnerError: err,
		Details:    make(map[string]any),
	}
}

// WithContext adds context to the error
func (e *AppError) WithContext(ctx context.Context) *AppError {
	e.Context = ctx
	return e
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]any) *AppError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// Mark sets the error code
func (e *AppError) Mark(code ErrorCode) *AppError {
	e.Code = code
	return e
}

// WithMessage sets the error code and message
func (e *AppError) WithMessage(message string) *AppError {
	e.Message = message
	return e
}

// NewErrorWithContext creates a new AppError with context
func NewErrorWithContext(ctx context.Context, code ErrorCode, err error) *AppError {
	ierr := &AppError{
		Code:       code,
		InnerError: err,
		Context:    ctx,
	}

	_, file, line, _ := runtime.Caller(1)
	logger.GetLoggerFromContext(ctx).Warnf(
		"[%s] [%s:%d] error: %w",
		code,
		file,
		line,
		ierr.Error(),
	)
	return ierr
}

// NewWarningWithContext creates a new AppError with context
func NewWarningWithContext(ctx context.Context, code ErrorCode, message string) *AppError {
	ierr := &AppError{
		Code:    code,
		Message: message,
		Context: ctx,
	}

	_, file, line, _ := runtime.Caller(1)
	logger.GetLoggerFromContext(ctx).Warnf(
		"[%s] [%s:%d] warning: %s",
		code,
		file,
		line,
		message,
	)
	return ierr
}
