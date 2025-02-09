package handlers

import (
	"net/http"

	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"date_of_birth"`
	StreetLine1 string `json:"street_line1"`
	StreetLine2 string `json:"street_line2"`
	City        string `json:"city"`
	Province    string `json:"province"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
}

// Register handles user registration.
// @Summary Register a new user
// @Description Register a new user with the provided details
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration details"
// @Success 200 {object} proto.RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/register [post]
func Register(authClient grpc.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		// Call the gRPC service
		resp, err := authClient.Register(c.Request.Context(), &proto.RegisterRequest{
			Username:    req.Username,
			Email:       req.Email,
			Password:    req.Password,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			Phone:       req.Phone,
			DateOfBirth: req.DateOfBirth,
			StreetLine1: req.StreetLine1,
			StreetLine2: req.StreetLine2,
			City:        req.City,
			Province:    req.Province,
			PostalCode:  req.PostalCode,
			Country:     req.Country,
		})
		if err != nil && resp == nil {
			utils.Error("Failed to register user", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to register user"})
			return
		}

		if !resp.Success {
			utils.Error("Failed to register user", map[string]interface{}{
				"error": resp.Message,
			})
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Message})
			return
		}

		// Return the response
		c.JSON(http.StatusOK, resp)
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles user login.
// @Summary Login
// @Description Login with the provided email/username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login details"
// @Success 200 {object} proto.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/login [post]
func Login(authClient grpc.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.Error("Failed to bind JSON", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		// Call the gRPC service
		resp, err := authClient.Login(c.Request.Context(), &proto.LoginRequest{
			Email:    req.Email,
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil && resp == nil {
			utils.Error("Failed to login", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to login"})
			return
		}

		if !resp.Success {
			utils.Error("Failed to login", map[string]interface{}{
				"error": resp.Message,
			})
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Message})
			return
		}

		// Return the response
		c.JSON(http.StatusOK, resp)
	}
}
