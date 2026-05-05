package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorDetail struct {
	Field string `json:"field,omitempty"`
	Issue string `json:"issue,omitempty"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details []ErrorDetail  `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type HealthzStatus struct {
	Service  string `json:"service"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

type HealthzResponse struct {
	Status   string          `json:"status"`
	Services []HealthzStatus `json:"services"`
}

func WriteJSON(c *gin.Context, status int, v any) {
	c.JSON(status, v)
}

// WriteError sends a standardized error envelope matching the api contract.
// optional field level details can be passed for validation errors.
func WriteError(c *gin.Context, status int, code, message string, details ...ErrorDetail) {
	body := ErrorEnvelope{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	}
	if len(details) > 0 {
		body.Error.Details = details
	}
	c.JSON(status, body)
}

// WriteHealthz sends a health check response with per-service status.
// if any dependency is down the overall status is set to degraded.
func WriteHealthz(c *gin.Context, status int, services []HealthzStatus) {
	overall := "ok"
	if status != http.StatusOK {
		overall = "degraded"
	}
	c.JSON(status, HealthzResponse{
		Status:   overall,
		Services: services,
	})
}
