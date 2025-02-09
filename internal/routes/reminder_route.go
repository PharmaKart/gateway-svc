package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/gin-gonic/gin"
)

func RegisterReminderRoutes(r *gin.RouterGroup, authClient grpc.AuthClient, reminderClient grpc.ReminderClient) {
}
