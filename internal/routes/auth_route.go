package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.RouterGroup, cfg *config.Config, authClient grpc.AuthClient) {
	r.POST("/auth/register", handlers.Register(authClient))
	r.POST("/auth/login", handlers.Login(authClient))
}
