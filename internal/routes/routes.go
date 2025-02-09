package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all routes for the application.
// @title PharmaKart Gateway API
// @version 1.0
// @description This is the API documentation for the PharmaKart Gateway Service.
// @termsOfService http://swagger.io/terms/
// @contact.name Ashutosh Sharma
// @contact.email asrma.sharma@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
func RegisterRoutes(r *gin.Engine, cfg *config.Config, authClient grpc.AuthClient, productClient grpc.ProductClient, orderClient grpc.OrderClient, paymentClient grpc.PaymentClient, reminderClient grpc.ReminderClient) {
	api := r.Group("/api/v1")
	// Register auth routes
	RegisterAuthRoutes(api, cfg, authClient)

	// Register product routes
	RegisterProductRoutes(api, cfg, authClient, productClient)

	// Register order routes
	RegisterOrderRoutes(api, cfg, authClient, orderClient)

	// Register payment routes
	RegisterPaymentRoutes(api, cfg, authClient, paymentClient)

	// Register reminder routes
	RegisterReminderRoutes(api, cfg, authClient, reminderClient)

	// Register health check route
	r.GET("/health", handlers.HealthCheck)
}
