package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.RouterGroup, authClient grpc.AuthClient) {
	r.POST("/register", handlers.Register(authClient))
	r.POST("/login", handlers.Login(authClient))
}
