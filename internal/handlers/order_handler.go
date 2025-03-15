package handlers

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"github.com/gin-gonic/gin"
)

type OrderItem struct {
	ProductID   string `json:"product_id" form:"product_id" binding:"required"`
	ProductName string `json:"product_name" form:"product_name" binding:"required"`
	Quantity    int    `json:"quantity" form:"quantity" binding:"required"`
}

type OrderRequest struct {
	Items []OrderItem `form:"items"`
}

type Order struct {
	OrderRequest
	Prescription *multipart.FileHeader `form:"prescription" swaggerignore:"true"`
}

// ErrorResponse represents an error response from the API
// @Description Error response
type ErrorResponse struct {
	Type    string            `json:"type" example:"VALIDATION_ERROR"`
	Message string            `json:"message" example:"Invalid request format"`
	Details map[string]string `json:"details,omitempty" example:"field:error message"`
}

// @Description Order placement request
type SwaggerOrderRequest struct {
	Items        []OrderItem `json:"items"`
	Prescription string      `json:"prescription" format:"binary"`
}

// PlaceOrder creates a new order
// @Summary Place a new order
// @Description Creates new order with the given product ID and quantity
// @Tags Orders
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param items formData string true "Order Items JSON"
// @Param prescription formData file false "Prescription Image"
// @Success 200 {object} proto.PlaceOrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [post]
func PlaceOrder(cfg *config.Config, orderClient grpc.OrderClient) gin.HandlerFunc {
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
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "Admins cannot place orders",
			})
			return
		}

		var req Order

		// Get the items JSON string from form data
		itemsStr := c.PostForm("items")

		// Create a temporary struct to unmarshal the JSON
		var tempRequest struct {
			Items []OrderItem `json:"items"`
		}

		// Unmarshal the JSON string into the temporary struct
		if err := json.Unmarshal([]byte(itemsStr), &tempRequest); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: map[string]string{"format": err.Error()},
			})
			return
		}

		// Check if items are provided
		if len(tempRequest.Items) == 0 {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: map[string]string{"items": "At least one item is required"},
			})
			return
		}

		// Assign the parsed items to req
		req.Items = tempRequest.Items

		// Handle prescription file separately
		file, _ := c.FormFile("prescription")
		req.Prescription = file

		var prescriptionURL *string

		// Check if a prescription is provided
		if req.Prescription != nil {
			// Validate file type
			allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
			ext := filepath.Ext(req.Prescription.Filename)
			if !allowedExtensions[ext] {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{
					Type:    "VALIDATION_ERROR",
					Message: "Invalid file format",
					Details: map[string]string{"format": "Only JPG, JPEG, PNG, and PDF files are allowed"},
				})
				return
			}

			// Upload prescription to S3
			url, err := utils.UploadImageToS3(c, cfg, "prescriptions", req.Prescription)
			if err != nil {
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
					Type:    "INTERNAL_ERROR",
					Message: "Failed to upload prescription",
				})
				return
			}

			prescriptionURL = &url
		}

		// Convert order items to gRPC format
		orderItems := make([]*proto.OrderItem, len(req.Items))
		for i, item := range req.Items {
			orderItems[i] = &proto.OrderItem{
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    int32(item.Quantity),
			}
		}

		// Call the gRPC service
		resp, err := orderClient.PlaceOrder(c.Request.Context(), &proto.PlaceOrderRequest{
			CustomerId:      customerID.(string),
			Items:           orderItems,
			PrescriptionUrl: prescriptionURL,
		})
		if err != nil {
			utils.Error("Failed to place order", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to place order",
			})
			return
		}

		// Check if the response indicates a failure
		if !resp.Success {
			utils.Error("Failed to place order", map[string]interface{}{
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
				Message: "Failed to place order",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// GenerateNewPaymentUrl generates a new payment URL for an order
// @Summary Generate a new payment URL
// @Description Generates a new payment URL for an order
// @Tags Orders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Success 200 {object} proto.GeneratePaymentURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id}/payment [post]
func GenerateNewPaymentUrl(orderClient grpc.OrderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		var customerID string
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User Role not found in token",
			})
			return
		}

		userId, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		customerID = userId.(string)
		if userRole == "admin" {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "Admins cannot generate payment URLs for customers",
			})
			return
		}

		orderID := c.Param("id")

		resp, err := orderClient.GenerateNewPaymentUrl(c.Request.Context(), &proto.GenerateNewPaymentUrlRequest{
			OrderId:    orderID,
			CustomerId: customerID,
		})
		if err != nil {
			utils.Error("Failed to generate payment URL", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to generate payment URL",
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to generate payment URL", map[string]interface{}{
				"error": resp,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to generate payment URL",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// GetOrder retrieves an order by ID
// @Summary Get an order
// @Description Retrieves an order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Success 200 {object} proto.GetOrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id} [get]
func GetOrder(orderClient grpc.OrderClient, paymentClient grpc.PaymentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		var customerID string
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User Role not found in token",
			})
			return
		}

		userId, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		customerID = userId.(string)
		if userRole == "admin" {
			customerID = "admin"
		}

		orderID := c.Param("id")

		orderResp, err := orderClient.GetOrder(c.Request.Context(), &proto.GetOrderRequest{
			OrderId:    orderID,
			CustomerId: customerID,
		})
		if err != nil {
			utils.Error("Failed to get order", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get order",
			})
			return
		}

		// Check if the response indicates a failure
		if !orderResp.Success {
			utils.Error("Failed to get order", map[string]interface{}{
				"error": orderResp,
			})

			if orderResp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(orderResp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			// Fallback if error structure is not available
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to get order",
			})
			return
		}
		orderData, err := json.Marshal(orderResp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal order data"})
			return
		}

		var response map[string]interface{}
		if err := json.Unmarshal(orderData, &response); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal order data"})
			return
		}

		// Append the payment status
		paymentResp, err := paymentClient.GetPaymentByOrderID(c.Request.Context(), &proto.GetPaymentByOrderIDRequest{
			OrderId:    orderID,
			CustomerId: customerID,
		})

		if err == nil && paymentResp.Success {
			response["payment_status"] = paymentResp.Status
		}

		c.JSON(http.StatusOK, response)
	}
}

// ListCustomersOrders retrieves all orders for a customer
// @Summary List all orders
// @Description Retrieves all orders for a customer
// @Tags Orders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page number"
// @Param limit query int false "Page limit"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param filter_column query string false "Filter column"
// @Param filter_operator query string false "Filter operator"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListCustomersOrdersResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [get]
func ListCustomersOrders(orderClient grpc.OrderClient) gin.HandlerFunc {
	return func(c *gin.Context) { // Get customer ID from the token
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

		resp, err := orderClient.ListCustomersOrders(c.Request.Context(), &proto.ListCustomersOrdersRequest{
			CustomerId: customerID.(string),
			Filter:     filter,
			SortBy:     sortBy,
			SortOrder:  sortOrder,
			Page:       int32(page),
			Limit:      int32(limit),
		})
		if err != nil {
			utils.Error("Failed to list orders", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to list orders",
			})
			return
		}

		// Check if the response indicates a failure
		if !resp.Success {
			utils.Error("Failed to list orders", map[string]interface{}{
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
				Message: "Failed to list orders",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// ListAllOrders retrieves all orders
// @Summary List all orders
// @Description Retrieves all orders
// @Tags Orders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page number"
// @Param limit query int false "Page limit"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param filter_column query string false "Filter column"
// @Param filter_operator query string false "Filter operator"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListAllOrdersResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/orders [get]
func ListAllOrders(orderClient grpc.OrderClient) gin.HandlerFunc {
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

		resp, err := orderClient.ListAllOrders(c.Request.Context(), &proto.ListAllOrdersRequest{
			Filter:    filter,
			SortBy:    sortBy,
			SortOrder: sortOrder,
			Page:      int32(page),
			Limit:     int32(limit),
		})
		if err != nil {
			utils.Error("Failed to list orders", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to list orders",
			})
			return
		}

		// Check if the response indicates a failure
		if !resp.Success {
			utils.Error("Failed to list all orders", map[string]interface{}{
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
				Message: "Failed to list orders",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

type OrderStatusRequest struct {
	Status string `json:"status"`
}

// UpdateOrder updates an order by ID
// @Summary Update an order
// @Description Updates an order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Param request body OrderStatusRequest true "Order Details"
// @Success 200 {object} proto.UpdateOrderStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/orders/{id} [put]
func UpdateOrderStatus(orderClient grpc.OrderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		var customerID string
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User Role not found in token",
			})
			return
		}

		userId, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		customerID = userId.(string)
		if userRole == "admin" {
			customerID = "admin"
		}

		orderID := c.Param("id")

		var req OrderStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: map[string]string{"format": err.Error()},
			})
			return
		}

		resp, err := orderClient.UpdateOrderStatus(c.Request.Context(), &proto.UpdateOrderStatusRequest{
			OrderId:    orderID,
			CustomerId: customerID,
			Status:     req.Status,
		})
		if err != nil {
			utils.Error("Failed to update order status", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to update order",
			})
			return
		}

		// Check if the response indicates a failure
		if !resp.Success {
			utils.Error("Failed to update order status", map[string]interface{}{
				"error": resp.Message,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			// Fallback if error structure is not available
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: resp.Message,
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
