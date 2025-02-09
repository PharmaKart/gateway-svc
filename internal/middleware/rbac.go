package middleware

import (
	"net/http"

	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"github.com/gin-gonic/gin"
)

func RBACMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			utils.Error("User not authenticated", map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		allowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			utils.Error("User not authorized", map[string]interface{}{
				"path": c.Request.URL.Path,
			})
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User not authorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
