package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/PharmaKart/gateway-svc/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterOrderRoutes(r *gin.RouterGroup, authClient grpc.AuthClient, orderClient grpc.OrderClient) {
	r.Use(middleware.AuthMiddleware(authClient))
	{
		r.POST("/orders", handlers.PlaceOrder(orderClient))
		r.GET("/orders", handlers.ListCustomersOrders(orderClient))
		r.GET("/orders/:id", handlers.GetOrder(orderClient))
		r.PUT("/orders/:id", handlers.UpdateOrderStatus(orderClient))
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(authClient))
	admin.Use(middleware.RBACMiddleware("admin"))
	{
		admin.GET("/orders", handlers.ListAllOrders(orderClient))
		admin.GET("/orders/:id", handlers.GetOrder(orderClient))
		admin.PUT("/orders/:id", handlers.UpdateOrderStatus(orderClient))
	}
}
