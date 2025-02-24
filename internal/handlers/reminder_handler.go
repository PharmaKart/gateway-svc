package handlers

import (
	"context"
	"net/http"

	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"github.com/gin-gonic/gin"
)

type ScheduleReminderRequest struct {
	CustomerID   string `json:"customer_id" binding:"required"`
	OrderID      string `json:"order_id" binding:"required"`
	ProductID    string `json:"product_id" binding:"required"`
	ReminderDate string `json:"reminder_date" binding:"required"`
}

// ScheduleReminder is a function that schedules a reminder for a customer
// @Summary Schedule a reminder
// @Description Schedule a reminder for a customer
// @Tags Reminders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param request body ScheduleReminderRequest true "Reminder Details"
// @Success 200 {object} proto.ScheduleReminderResponse
// @Router /api/v1/reminders [post]
func ScheduleReminder(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req proto.ScheduleReminderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := reminderClient.ScheduleReminder(context.Background(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule reminder: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// ListReminders lists all reminders
// @Summary List reminders
// @Description List all reminders
// @Tags Reminders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param filter query string false "Filter field"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListRemindersResponse
// @Router /api/v1/admin/reminders [get]
func ListReminders(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := utils.GetIntQueryParam(c, "page", 1)
		limit := utils.GetIntQueryParam(c, "limit", 10)
		sortBy := c.Query("sort_by")
		sortOrder := c.Query("sort_order")
		filter := c.Query("filter")
		filterValue := c.Query("filter_value")

		resp, err := reminderClient.ListReminders(context.Background(), &proto.ListRemindersRequest{
			Page:        int32(page),
			Limit:       int32(limit),
			SortBy:      sortBy,
			SortOrder:   sortOrder,
			Filter:      filter,
			FilterValue: filterValue,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reminders: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// ListCustomerReminders lists all reminders for a customer
// @Summary List customer reminders
// @Description List all reminders for a customer
// @Tags Reminders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param customer_id query string true "Customer ID"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param filter query string false "Filter field"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListRemindersResponse
// @Router /api/v1/reminders [get]
func ListCustomerReminders(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
			return
		}

		page := utils.GetIntQueryParam(c, "page", 1)
		limit := utils.GetIntQueryParam(c, "limit", 10)
		sortBy := c.Query("sort_by")
		sortOrder := c.Query("sort_order")
		filter := c.Query("filter")
		filterValue := c.Query("filter_value")

		resp, err := reminderClient.ListCustomerReminders(context.Background(), &proto.ListCustomerRemindersRequest{
			CustomerId:  customerID.(string),
			Page:        int32(page),
			Limit:       int32(limit),
			SortBy:      sortBy,
			SortOrder:   sortOrder,
			Filter:      filter,
			FilterValue: filterValue,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reminders: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// DeleteReminder deletes a reminder
// @Summary Delete a reminder
// @Description Deletes a reminder
// @Tags Reminders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param reminder_id path string true "Reminder ID"
// @Success 200 {object} proto.DeleteReminderResponse
// @Router /api/v1/reminders/{reminder_id} [delete]
func DeleteReminder(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
			return
		}

		reminderID := c.Param("reminder_id")

		resp, err := reminderClient.DeleteReminder(context.Background(), &proto.DeleteReminderRequest{
			CustomerId: customerID.(string),
			ReminderId: reminderID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reminder: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// UpdateReminder updates a reminder
// @Summary Update a reminder
// @Description Updates a reminder
// @Tags Reminders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param reminder_id path string true "Reminder ID"
// @Param request body ScheduleReminderRequest true "Reminder Details"
// @Success 200 {object} proto.UpdateReminderResponse
// @Router /api/v1/reminders/{reminder_id} [put]
func UpdateReminder(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
			return
		}

		reminderID := c.Param("reminder_id")

		var req proto.UpdateReminderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		resp, err := reminderClient.UpdateReminder(context.Background(), &proto.UpdateReminderRequest{
			CustomerId:   customerID.(string),
			ReminderId:   reminderID,
			OrderId:      req.OrderId,
			ReminderDate: req.ReminderDate,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reminder: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// ToggleReminder toggles a reminder
// @Summary Toggle a reminder
// @Description Toggles a reminder
// @Tags Reminders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param reminder_id path string true "Reminder ID"
// @Success 200 {object} proto.ToggleReminderResponse
// @Router /api/v1/reminders/{reminder_id} [patch]
func ToggleReminder(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
			return
		}

		reminderID := c.Param("reminder_id")

		resp, err := reminderClient.ToggleReminder(context.Background(), &proto.ToggleReminderRequest{
			CustomerId: customerID.(string),
			ReminderId: reminderID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle reminder: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// ListReminderLogs lists all reminder logs
// @Summary List reminder logs
// @Description List all reminder logs
// @Tags Reminders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param reminder_id path string true "Reminder ID"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param filter query string false "Filter field"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListReminderLogsResponse
// @Router /api/v1/reminders/{reminder_id}/logs [get]
func ListReminderLogs(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User Role not found in token"})
			return
		}

		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
			return
		}

		if userRole == "admin" {
			customerID = "admin"
		}

		reminderID := c.Param("reminder_id")
		page := utils.GetIntQueryParam(c, "page", 1)
		limit := utils.GetIntQueryParam(c, "limit", 10)
		sortBy := c.Query("sort_by")
		sortOrder := c.Query("sort_order")
		filter := c.Query("filter")
		filterValue := c.Query("filter_value")

		resp, err := reminderClient.ListReminderLogs(context.Background(), &proto.ListReminderLogsRequest{
			CustomerId:  customerID.(string),
			ReminderId:  reminderID,
			Page:        int32(page),
			Limit:       int32(limit),
			SortBy:      sortBy,
			SortOrder:   sortOrder,
			Filter:      filter,
			FilterValue: filterValue,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reminder logs: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
