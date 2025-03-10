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
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/reminders [post]
func ScheduleReminder(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req proto.ScheduleReminderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: map[string]string{"format": err.Error()},
			})
			return
		}

		resp, err := reminderClient.ScheduleReminder(context.Background(), &req)
		if err != nil {
			utils.Error("Failed to schedule reminder", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to schedule reminder",
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to schedule reminder", map[string]interface{}{
				"error": resp,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			// Fallback if error structure is not available
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to schedule reminder",
			})
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
// @Param filter_column query string false "Filter column"
// @Param filter_operator query string false "Filter operator"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListRemindersResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/admin/reminders [get]
func ListReminders(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		sortBy := c.Query("sort_by")
		sortOrder := c.Query("sort_order")
		page := utils.GetIntQueryParam(c, "page", 1)
		limit := utils.GetIntQueryParam(c, "limit", 0)

		column := c.Query("filter_column")
		operator := c.Query("filter_operator")
		value := c.Query("filter_value")

		var filter *proto.Filter
		if column != "" && operator != "" && value != "" {
			filter = &proto.Filter{
				Column:   column,
				Operator: operator,
				Value:    value,
			}
		}

		resp, err := reminderClient.ListReminders(context.Background(), &proto.ListRemindersRequest{
			Filter:    filter,
			SortBy:    sortBy,
			SortOrder: sortOrder,
			Page:      int32(page),
			Limit:     int32(limit),
		})
		if err != nil {
			utils.Error("Failed to get reminders", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get reminders",
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to get reminders", map[string]interface{}{
				"error": resp,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to get reminders",
			})
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
// @Param filter_column query string false "Filter column"
// @Param filter_operator query string false "Filter operator"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListRemindersResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/reminders [get]
func ListCustomerReminders(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		sortBy := c.Query("sort_by")
		sortOrder := c.Query("sort_order")
		page := utils.GetIntQueryParam(c, "page", 1)
		limit := utils.GetIntQueryParam(c, "limit", 0)

		column := c.Query("filter_column")
		operator := c.Query("filter_operator")
		value := c.Query("filter_value")

		var filter *proto.Filter
		if column != "" && operator != "" && value != "" {
			filter = &proto.Filter{
				Column:   column,
				Operator: operator,
				Value:    value,
			}
		}

		resp, err := reminderClient.ListCustomerReminders(context.Background(), &proto.ListCustomerRemindersRequest{
			CustomerId: customerID.(string),
			Filter:     filter,
			SortBy:     sortBy,
			SortOrder:  sortOrder,
			Page:       int32(page),
			Limit:      int32(limit),
		})
		if err != nil {
			utils.Error("Failed to get reminders", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get reminders",
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to get customer reminders", map[string]interface{}{
				"error": resp,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to get reminders",
			})
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
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/reminders/{reminder_id} [delete]
func DeleteReminder(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		reminderID := c.Param("reminder_id")

		resp, err := reminderClient.DeleteReminder(context.Background(), &proto.DeleteReminderRequest{
			CustomerId: customerID.(string),
			ReminderId: reminderID,
		})
		if err != nil {
			utils.Error("Failed to delete reminder", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to delete reminder",
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to delete reminder", map[string]interface{}{
				"error": resp.Message,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: resp.Message,
			})
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
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/reminders/{reminder_id} [put]
func UpdateReminder(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		reminderID := c.Param("reminder_id")

		var req proto.UpdateReminderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: map[string]string{"format": err.Error()},
			})
			return
		}

		req.CustomerId = customerID.(string)
		req.ReminderId = reminderID

		resp, err := reminderClient.UpdateReminder(context.Background(), &req)
		if err != nil {
			utils.Error("Failed to update reminder", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to update reminder",
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to update reminder", map[string]interface{}{
				"error": resp.Message,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: resp.Message,
			})
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
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/reminders/{reminder_id} [patch]
func ToggleReminder(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		reminderID := c.Param("reminder_id")

		resp, err := reminderClient.ToggleReminder(context.Background(), &proto.ToggleReminderRequest{
			CustomerId: customerID.(string),
			ReminderId: reminderID,
		})
		if err != nil {
			utils.Error("Failed to toggle reminder", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to toggle reminder",
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to toggle reminder", map[string]interface{}{
				"error": resp.Message,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: resp.Message,
			})
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
// @Param filter_column query string false "Filter column"
// @Param filter_operator query string false "Filter operator"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListReminderLogsResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/reminders/{reminder_id}/logs [get]
func ListReminderLogs(reminderClient grpc.ReminderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User Role not found in token",
			})
			return
		}

		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		if userRole == "admin" {
			customerID = "admin"
		}

		reminderID := c.Param("reminder_id")

		sortBy := c.Query("sort_by")
		sortOrder := c.Query("sort_order")
		page := utils.GetIntQueryParam(c, "page", 1)
		limit := utils.GetIntQueryParam(c, "limit", 0)

		column := c.Query("filter_column")
		operator := c.Query("filter_operator")
		value := c.Query("filter_value")

		var filter *proto.Filter
		if column != "" && operator != "" && value != "" {
			filter = &proto.Filter{
				Column:   column,
				Operator: operator,
				Value:    value,
			}
		}

		resp, err := reminderClient.ListReminderLogs(context.Background(), &proto.ListReminderLogsRequest{
			CustomerId: customerID.(string),
			ReminderId: reminderID,
			Filter:     filter,
			SortBy:     sortBy,
			SortOrder:  sortOrder,
			Page:       int32(page),
			Limit:      int32(limit),
		})
		if err != nil {
			utils.Error("Failed to get reminder logs", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get reminder logs",
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to get reminder logs", map[string]interface{}{
				"error": resp,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to get reminder logs",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
