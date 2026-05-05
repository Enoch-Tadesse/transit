package domain

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrConflict      = errors.New("resource conflict")
	ErrForbidden     = errors.New("forbidden")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrValidation    = errors.New("validation error")
	ErrUnprocessable = errors.New("unprocessable entity")
	ErrInternal      = errors.New("internal error")
)
