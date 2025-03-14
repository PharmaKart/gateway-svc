package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/PharmaKart/gateway-svc/internal/middleware"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/gin-gonic/gin"
)

func RegisterProductRoutes(r *gin.RouterGroup, cfg *config.Config, authClient grpc.AuthClient, productClient grpc.ProductClient) {
	r.GET("/products", handlers.GetProducts(productClient))
	r.GET("/products/:id", handlers.GetProduct(productClient))

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(authClient))
	admin.Use(middleware.RBACMiddleware("admin"))
	{
		admin.POST("/products", handlers.CreateProduct(cfg, productClient))
		admin.PUT("/products/:id", handlers.UpdateProduct(cfg, productClient))
		admin.DELETE("/products/:id", handlers.DeleteProduct(productClient))
		admin.PUT("/products/:id/stock", handlers.UpdateStock(productClient))
		admin.GET("/products/:id/logs", handlers.GetInventoryLogs(productClient))
	}
}
