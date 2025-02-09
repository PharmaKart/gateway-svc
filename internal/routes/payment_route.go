package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/gin-gonic/gin"
)

func RegisterPaymentRoutes(r *gin.RouterGroup, cfg *config.Config, authClient grpc.AuthClient, paymentClient grpc.PaymentClient) {
}
