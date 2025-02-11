package handlers

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"github.com/gin-gonic/gin"
)

type OrderItem struct {
	ProductID   string `json:"product_id" form:"product_id"`
	ProductName string `json:"product_name" form:"product_name"`
	Quantity    int    `json:"quantity" form:"quantity"`
}

type OrderRequest struct {
	Items []OrderItem `form:"items"`
}

type Order struct {
	OrderRequest
	Prescription *multipart.FileHeader `form:"prescription" swaggerignore:"true"`
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
// @Router /api/v1/orders [post]
func PlaceOrder(orderClient grpc.OrderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
			return
		}

		var req Order

		// Get the items JSON string from form data
		itemsStr := c.PostForm("items")
		fmt.Printf("Received items string: %s\n", itemsStr)

		// Create a temporary struct to unmarshal the JSON
		var tempRequest struct {
			Items []OrderItem `json:"items"`
		}

		// Unmarshal the JSON string into the temporary struct
		if err := json.Unmarshal([]byte(itemsStr), &tempRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse items: " + err.Error()})
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
			allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
			ext := filepath.Ext(req.Prescription.Filename)
			if !allowedExtensions[ext] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only .jpg, .jpeg, and .png are allowed"})
				return
			}

			// Upload prescription to S3
			url, err := utils.UploadImageToS3(c, "prescriptions", req.Prescription)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image to S3: " + err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
// @Router /api/v1/orders/{id} [get]

func GetOrder(orderClient grpc.OrderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		var customerID string
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User Role not found in token"})
			return
		}

		userId, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
			return
		}

		customerID = userId.(string)
		if userRole == "admin" {
			customerID = "admin"
		}

		orderID := c.Param("id")

		resp, err := orderClient.GetOrder(c.Request.Context(), &proto.GetOrderRequest{
			OrderId:    orderID,
			CustomerId: customerID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
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
// @Param filter query string false "Filter field"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListCustomersOrdersResponse
// @Router /api/v1/orders [get]
func ListCustomersOrders(orderClient grpc.OrderClient) gin.HandlerFunc {
	return func(c *gin.Context) { // Get customer ID from the token
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

		resp, err := orderClient.ListCustomersOrders(c.Request.Context(), &proto.ListCustomersOrdersRequest{
			CustomerId:  customerID.(string),
			Page:        int32(page),
			Limit:       int32(limit),
			SortBy:      sortBy,
			SortOrder:   sortOrder,
			Filter:      filter,
			FilterValue: filterValue,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
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
// @Param filter query string false "Filter field"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListAllOrdersResponse
// @Router /api/v1/admin/orders [get]
func ListAllOrders(orderClient grpc.OrderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := utils.GetIntQueryParam(c, "page", 1)
		limit := utils.GetIntQueryParam(c, "limit", 10)
		sortBy := c.Query("sort_by")
		sortOrder := c.Query("sort_order")
		filter := c.Query("filter")
		filterValue := c.Query("filter_value")
		resp, err := orderClient.ListAllOrders(c.Request.Context(), &proto.ListAllOrdersRequest{
			Page:        int32(page),
			Limit:       int32(limit),
			SortBy:      sortBy,
			SortOrder:   sortOrder,
			Filter:      filter,
			FilterValue: filterValue,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
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
// @Router /api/v1/orders/{id} [put]
func UpdateOrderStatus(orderClient grpc.OrderClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")

		var req OrderStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		resp, err := orderClient.UpdateOrderStatus(c.Request.Context(), &proto.UpdateOrderStatusRequest{
			OrderId: orderID,
			Status:  req.Status,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
