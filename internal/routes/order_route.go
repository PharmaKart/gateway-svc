package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/PharmaKart/gateway-svc/internal/middleware"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/gin-gonic/gin"
)

func RegisterOrderRoutes(r *gin.RouterGroup, cfg *config.Config, authClient grpc.AuthClient, orderClient grpc.OrderClient, paymentClient grpc.PaymentClient) {
	r.Use(middleware.AuthMiddleware(authClient))
	{
		r.POST("/orders", handlers.PlaceOrder(cfg, orderClient))
		r.GET("/orders", handlers.ListCustomersOrders(orderClient))
		r.GET("/orders/:id", handlers.GetOrder(orderClient, paymentClient))
		r.PUT("/orders/:id", handlers.UpdateOrderStatus(orderClient))
		r.POST("/orders/:id/payment", handlers.GenerateNewPaymentUrl(orderClient))
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(authClient))
	admin.Use(middleware.RBACMiddleware("admin"))
	{
		admin.GET("/orders", handlers.ListAllOrders(orderClient))
		admin.GET("/orders/:id", handlers.GetOrder(orderClient, paymentClient))
		admin.PUT("/orders/:id", handlers.UpdateOrderStatus(orderClient))
	}
}
