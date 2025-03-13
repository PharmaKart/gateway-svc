package main

import (
	"net/http"

	docs "github.com/PharmaKart/gateway-svc/docs"
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

	// Initialize gRPC client for product service
	productConn, err := grpc.NewClient(cfg.ProductServiceURL)
	if err != nil {
		utils.Logger.Fatal("Failed to connect to product service", map[string]interface{}{
			"error": err,
		})
	}

	productClient := grpc.NewProductServiceClient(productConn.Conn())
	defer productConn.Close()

	// Initialize gRPC client for order service
	orderConn, err := grpc.NewClient(cfg.OrderServiceURL)
	if err != nil {
		utils.Logger.Fatal("Failed to connect to order service", map[string]interface{}{
			"error": err,
		})
	}

	orderClient := grpc.NewOrderServiceClient(orderConn.Conn())
	defer orderConn.Close()

	// Initialize gRPC client for payment service
	paymentConn, err := grpc.NewClient(cfg.PaymentServiceURL)
	if err != nil {
		utils.Logger.Fatal("Failed to connect to payment service", map[string]interface{}{
			"error": err,
		})
	}

	paymentClient := grpc.NewPaymentServiceClient(paymentConn.Conn())
	defer paymentConn.Close()

	// Initialize gRPC client for reminder service
	reminderConn, err := grpc.NewClient(cfg.ReminderServiceURL)
	if err != nil {
		utils.Logger.Fatal("Failed to connect to reminder service", map[string]interface{}{
			"error": err,
		})
	}

	reminderClient := grpc.NewReminderServiceClient(reminderConn.Conn())
	defer reminderConn.Close()

	// Set to Release mode once in production
	gin.SetMode(gin.ReleaseMode)

	// Initialize Gin router
	r := gin.Default()

	// Add Swagger documentation
	docs.SwaggerInfo.Title = "PharmaKart Gateway API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "This is the API documentation for the PharmaKart Gateway Service."

	// Set CORS headers
	r.Use(utils.NewCors())

	// Redirect /swagger to /swagger/index.html
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	// Add Swagger endpoint
	r.GET("/swagger/*any", SwaggerAuthMiddleware(), ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1),
	)) // Register auth routes
	routes.RegisterRoutes(r, cfg, authClient, productClient, orderClient, paymentClient, reminderClient)

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
