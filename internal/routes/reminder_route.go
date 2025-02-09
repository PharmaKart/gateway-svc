package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/gin-gonic/gin"
)

func RegisterReminderRoutes(r *gin.RouterGroup, cfg *config.Config, authClient grpc.AuthClient, reminderClient grpc.ReminderClient) {
}
