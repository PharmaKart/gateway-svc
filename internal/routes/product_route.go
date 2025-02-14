package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/PharmaKart/gateway-svc/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterProductRoutes(r *gin.RouterGroup, authClient grpc.AuthClient, productClient grpc.ProductClient) {
	r.GET("/products", handlers.GetProducts(productClient))
	r.GET("/products/:id", handlers.GetProduct(productClient))

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(authClient))
	admin.Use(middleware.RBACMiddleware("admin"))
	{
		admin.POST("/products", handlers.CreateProduct(productClient))
		admin.PUT("/products/:id", handlers.UpdateProduct(productClient))
		admin.DELETE("/products/:id", handlers.DeleteProduct(productClient))
		admin.PUT("/products/:id/stock", handlers.UpdateStock(productClient))
	}
}
