package response

import (
	"errors"
	"net/http"

	"github.com/henok/transit-backend/internal/domain"
)

type HTTPStatus struct {
	Status int
	Code   string
}

// maps sentinel domain errors to their http status and error.code per the contract.
// anything not in this map falls through to 500.
var domainToHTTP = map[error]HTTPStatus{
	domain.ErrValidation:    {Status: http.StatusBadRequest, Code: "validation_error"},
	domain.ErrUnauthorized:  {Status: http.StatusUnauthorized, Code: "unauthorized"},
	domain.ErrForbidden:     {Status: http.StatusForbidden, Code: "forbidden"},
	domain.ErrNotFound:      {Status: http.StatusNotFound, Code: "not_found"},
	domain.ErrConflict:      {Status: http.StatusConflict, Code: "conflict"},
	domain.ErrUnprocessable: {Status: http.StatusUnprocessableEntity, Code: "unprocessable"},
}

// MapError converts a domain sentinel error to an http status code,
// error code string, and user facing message. unrecognized errors
// fall through to a generic 500 internal_error.
func MapError(err error) (int, string, string) {
	if err == nil {
		return http.StatusOK, "", ""
	}

	for domainErr, mapping := range domainToHTTP {
		if errors.Is(err, domainErr) {
			return mapping.Status, mapping.Code, err.Error()
		}
	}

	return http.StatusInternalServerError, "internal_error", "An unexpected error occurred."
}
