package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
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
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
