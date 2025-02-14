package middleware

import (
	"net/http"
	"strings"

	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authClient grpc.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.URL.Path == "/api/v1/payment/webhook" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error("Authorization header is missing", map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.Error("Invalid authorization header", map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}
		token := tokenParts[1]

		resp, err := authClient.VerifyToken(c.Request.Context(), &proto.VerifyTokenRequest{Token: token})
		if err != nil {
			utils.Error("Failed to verify token", map[string]interface{}{
				"error": err,
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			c.Abort()
			return
		}

		if !resp.Success {
			utils.Error(resp.Message, map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": resp.Message})
			c.Abort()
			return
		}

		c.Set("user_id", resp.UserId)
		c.Set("user_role", resp.Role)

		utils.Info("User authenticated", map[string]interface{}{
			"user_id":   resp.UserId,
			"user_role": resp.Role,
		})

		c.Next()
	}
}
