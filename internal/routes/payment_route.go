package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/gin-gonic/gin"
)

func RegisterPaymentRoutes(r *gin.RouterGroup, authClient grpc.AuthClient, paymentClient grpc.PaymentClient) {
}
