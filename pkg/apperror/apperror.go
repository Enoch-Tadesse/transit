package apperror

import "fmt"

// AppError carries a machine readable error code and optional field level detail
// so the http layer can build the exact envelope shape from api contract without
// having to reparse the error string.
type AppError struct {
	Code       string
	Message    string
	Field      string
	FieldIssue string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) HasField() bool {
	return e.Field != "" || e.FieldIssue != ""
}

func NewValidationError(msg string) *AppError {
	return &AppError{Code: "validation_error", Message: msg}
}

func NewValidationFieldError(field, issue string) *AppError {
	return &AppError{
		Code:       "validation_error",
		Message:    field + " " + issue,
		Field:      field,
		FieldIssue: issue,
	}
}

func NewUnauthorizedError(msg string) *AppError {
	return &AppError{Code: "unauthorized", Message: msg}
}

func NewForbiddenError(msg string) *AppError {
	return &AppError{Code: "forbidden", Message: msg}
}

func NewNotFoundError(msg string) *AppError {
	return &AppError{Code: "not_found", Message: msg}
}

func NewConflictError(msg string) *AppError {
	return &AppError{Code: "conflict", Message: msg}
}

func NewUnprocessableError(msg string) *AppError {
	return &AppError{Code: "unprocessable", Message: msg}
}

func NewInternalError(msg string) *AppError {
	return &AppError{Code: "internal_error", Message: msg}
}
