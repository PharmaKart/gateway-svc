package routes

import (
	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/handlers"
	"github.com/PharmaKart/gateway-svc/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterReminderRoutes(r *gin.RouterGroup, authClient grpc.AuthClient, reminderClient grpc.ReminderClient) {
	r.Use(middleware.AuthMiddleware(authClient))
	{
		r.POST("/reminders", handlers.ScheduleReminder(reminderClient))
		r.GET("/reminders", handlers.ListCustomerReminders(reminderClient))
		r.PUT("/reminders/:id", handlers.UpdateReminder(reminderClient))
		r.DELETE("/reminders/:id", handlers.DeleteReminder(reminderClient))
		r.PATCH("/reminders/:id", handlers.ToggleReminder(reminderClient))
		r.GET("/reminders/:id/logs", handlers.ListReminderLogs(reminderClient))
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(authClient))
	admin.Use(middleware.RBACMiddleware("admin"))
	{
		admin.GET("/reminders", handlers.ListReminders(reminderClient))
	}

}
