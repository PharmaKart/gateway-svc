package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/PharmaKart/gateway-svc/internal/middleware"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/gin-gonic/gin"
)

func RegisterPaymentRoutes(r *gin.RouterGroup, cfg *config.Config, authClient grpc.AuthClient, paymentClient grpc.PaymentClient) {
	r.POST("/payment/webhook", handlers.HandleWebhook(cfg, paymentClient))

	r.Use(middleware.AuthMiddleware(authClient))
	{
		r.GET("/payment/:id", handlers.GetPayment(paymentClient))
		r.GET("/payment/order/:id", handlers.GetPaymentByOrderID(paymentClient))
	}

}
