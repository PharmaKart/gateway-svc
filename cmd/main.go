package main

import (
	"net/http"

	_ "github.com/PharmaKart/gateway-svc/docs"
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/routes"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SwaggerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := "admin"
		password := "password"

		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != username || pass != password {
			c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

func main() {
	// Initialize logger
	utils.InitLogger()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize gRPC client for authentication service
	authConn, err := grpc.NewClient(cfg.AuthServiceURL)
	if err != nil {
		utils.Logger.Fatal("Failed to connect to authentication service", map[string]interface{}{
			"error": err,
		})
	}

	authClient := grpc.NewAuthServiceClient(authConn.Conn())
	defer authConn.Close()

	// Initialize Gin router
	r := gin.Default()

	// Add Swagger endpoint
	r.GET("/swagger/*any", SwaggerAuthMiddleware(), ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register auth routes
	routes.RegisterRoutes(r, cfg, authClient)

	// Start server
	utils.Info("Starting gateway service", map[string]interface{}{
		"port": cfg.Port,
	})
	if err := r.Run(":" + cfg.Port); err != nil {
		utils.Logger.Fatal("Failed to start server", map[string]interface{}{
			"error": err,
		})
	}
}
