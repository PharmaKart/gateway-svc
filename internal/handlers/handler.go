package handlers

import (
	"net/http"

	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Type    string            `json:"type"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Convert proto error to HTTP response
func convertProtoErrorToResponse(protoError *proto.Error) (ErrorResponse, int) {
	errorResp := ErrorResponse{
		Type:    protoError.Type,
		Message: protoError.Message,
	}

	// Convert details to map
	if len(protoError.Details) > 0 {
		errorResp.Details = make(map[string]string)
		for _, kv := range protoError.Details {
			errorResp.Details[kv.Key] = kv.Value
		}
	}

	// Determine status code based on error type
	statusCode := http.StatusInternalServerError
	switch protoError.Type {
	case "VALIDATION_ERROR":
		statusCode = http.StatusBadRequest
	case "AUTH_ERROR":
		statusCode = http.StatusUnauthorized
	case "NOT_FOUND_ERROR":
		statusCode = http.StatusNotFound
	case "CONFLICT_ERROR":
		statusCode = http.StatusConflict
	case "INTERNAL_ERROR":
		statusCode = http.StatusInternalServerError
	}

	return errorResp, statusCode
}

// HealthResponse represents the response for the health check endpoint.
// @Description Health check response
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// HealthCheck handles health check requests.
// @Summary Health check
// @Description Check if the service is running
// @Tags Utility
// @Produce json
// @Success 200 {object} HealthResponse
// @Error 500 {object} ErrorResponse
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "good"})
}
