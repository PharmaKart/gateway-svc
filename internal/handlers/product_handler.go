package handlers

import (
	"context"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/PharmaKart/gateway-svc/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ProductRequest struct {
	Name                 string  `json:"name" form:"name" binding:"required" example:"Paracetamol"`
	Description          string  `json:"description" form:"description" binding:"required" example:"Pain relief medication"`
	Price                float64 `json:"price" form:"price" binding:"required,gt=0" example:"9.99"`
	Stock                int32   `json:"stock" form:"stock" binding:"required,gt=0" example:"100"`
	RequiresPrescription bool    `json:"requires_prescription" form:"requires_prescription" example:"true"`
}

type ProductUpdate struct {
	Name                 string  `json:"name" form:"name" binding:"required" example:"Paracetamol"`
	Description          string  `json:"description" form:"description" binding:"required" example:"Pain relief medication"`
	Price                float64 `json:"price" form:"price" binding:"required,gt=0" example:"9.99"`
	RequiresPrescription bool    `json:"requires_prescription" form:"requires_prescription" example:"true"`
}

type Product struct {
	ProductRequest
	Image *multipart.FileHeader `form:"image" swaggerignore:"true"`
}

type UpdateProductReq struct {
	ProductUpdate
	Image *multipart.FileHeader `form:"image" swaggerignore:"true"`
}

// CreateProduct adds a new product to the inventory
// @Summary Add a new product
// @Description Adds a new product to the inventory
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param name formData string true "Product Name" example:"Paracetamol"
// @Param description formData string true "Product Description" example:"Pain relief medication"
// @Param price formData number true "Product Price" example:"9.99"
// @Param stock formData integer true "Stock Quantity" example:"100"
// @Param requires_prescription formData boolean false "Requires Prescription" example:"true"
// @Param image formData file false "Product Image"
// @Success 200 {object} proto.CreateProductResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 409 {object} utils.ErrorResponse "Conflict"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/admin/products [post]
func CreateProduct(cfg *config.Config, productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Product
		if err := c.ShouldBind(&req); err != nil {
			utils.Error("Failed to bind request", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: map[string]string{"format": err.Error()},
			})
			return
		}

		var imageURL string

		if req.Image != nil {
			// Validate file type
			allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
			ext := filepath.Ext(req.Image.Filename)
			if !allowedExtensions[ext] {
				utils.Error("Invalid file format", map[string]interface{}{
					"extension": ext,
				})
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{
					Type:    "VALIDATION_ERROR",
					Message: "Invalid file format",
					Details: map[string]string{"format": "Only JPG, JPEG, and PNG files are allowed"},
				})
				return
			}

			// Upload image to S3
			imageURLResp, err := utils.UploadImageToS3(c, cfg, "products", req.Image)
			if err != nil {
				utils.Error("Failed to upload image to S3", map[string]interface{}{
					"error": err,
				})
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
					Type:    "INTERNAL_ERROR",
					Message: "Failed to upload image",
					Details: map[string]string{"error": err.Error()},
				})
				return
			}
			imageURL = imageURLResp
		}

		// Call the gRPC service to create product
		resp, err := productClient.CreateProduct(context.Background(), &proto.CreateProductRequest{
			Product: &proto.Product{
				Name:                 req.Name,
				Description:          req.Description,
				Price:                req.Price,
				Stock:                int32(req.Stock),
				RequiresPrescription: req.RequiresPrescription,
				ImageUrl:             imageURL,
			},
		})

		if err != nil {
			utils.Error("Failed to create product", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to create product",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to create product", map[string]interface{}{
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
				Message: "Failed to create product",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// GetProduct fetches a product by ID
// @Summary Get a product
// @Description Fetches a product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} proto.GetProductResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/products/{id} [get]
func GetProduct(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		resp, err := productClient.GetProduct(context.Background(), &proto.GetProductRequest{
			ProductId: productID,
		})
		if err != nil {
			utils.Error("Failed to get product", map[string]interface{}{
				"error":      err,
				"product_id": productID,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get product",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to get product", map[string]interface{}{
				"error":      resp,
				"product_id": productID,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			// Fallback if error structure is not available
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to get product",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// GetProducts fetches a list of products
// @Summary Get all products
// @Description Fetches a list of products
// @Tags Products
// @Accept json
// @Produce json
// @Param page query integer false "Page number"
// @Param limit query integer false "Number of items per page"
// @Param sort_by query string false "Sort by column"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param filter_column query string false "Filter column"
// @Param filter_operator query string false "Filter operator"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListProductsResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/products [get]
func GetProducts(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := c.Query("search")
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

		resp, err := productClient.ListProducts(context.Background(), &proto.ListProductsRequest{
			Search:    search,
			Filter:    filter,
			SortBy:    sortBy,
			SortOrder: sortOrder,
			Page:      int32(page),
			Limit:     int32(limit),
		})

		if err != nil {
			utils.Error("Failed to get products", map[string]interface{}{
				"error": err,
				"page":  page,
				"limit": limit,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get products",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to get products", map[string]interface{}{
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
				Message: "Failed to get products",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// UpdateProduct updates a product by ID
// @Summary Update a product
// @Description Updates a product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param request body UpdateProductReq true "Product Details"
// @Success 200 {object} proto.UpdateProductResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/admin/products/{id} [put]
func UpdateProduct(cfg *config.Config, productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		var req UpdateProductReq
		if err := c.ShouldBind(&req); err != nil {
			utils.Error("Failed to bind request", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: map[string]string{"format": err.Error()},
			})
			return
		}

		var imageURL string

		if req.Image.Filename != "" {
			// Validate file type
			allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
			ext := filepath.Ext(req.Image.Filename)
			if !allowedExtensions[ext] {
				utils.Error("Invalid file format", map[string]interface{}{
					"extension": ext,
				})
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{
					Type:    "VALIDATION_ERROR",
					Message: "Invalid file format",
					Details: map[string]string{"format": "Only JPG, JPEG, and PNG files are allowed"},
				})
				return
			}

			// Upload image to S3
			imageURLResp, err := utils.UploadImageToS3(c, cfg, "products", req.Image)
			if err != nil {
				utils.Error("Failed to upload image to S3", map[string]interface{}{
					"error": err,
				})
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
					Type:    "INTERNAL_ERROR",
					Message: "Failed to upload image",
					Details: map[string]string{"error": err.Error()},
				})
				return
			}
			imageURL = imageURLResp
		}

		resp, err := productClient.UpdateProduct(context.Background(), &proto.UpdateProductRequest{
			ProductId: productID,
			Product: &proto.Product{
				Name:                 req.Name,
				Description:          req.Description,
				Price:                req.Price,
				RequiresPrescription: req.RequiresPrescription,
				ImageUrl:             imageURL,
			},
		})
		if err != nil {
			utils.Error("Failed to update product", map[string]interface{}{
				"error":      err,
				"product_id": productID,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to update product",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to update product", map[string]interface{}{
				"error":      resp.Message,
				"product_id": productID,
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

// DeleteProduct deletes a product by ID
// @Summary Delete a product
// @Description Deletes a product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Success 200 {object} proto.DeleteProductResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/admin/products/{id} [delete]
func DeleteProduct(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		resp, err := productClient.DeleteProduct(context.Background(), &proto.DeleteProductRequest{
			ProductId: productID,
		})
		if err != nil {
			utils.Error("Failed to delete product", map[string]interface{}{
				"error":      err,
				"product_id": productID,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to delete product",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to delete product", map[string]interface{}{
				"error":      resp.Message,
				"product_id": productID,
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

type StockRequest struct {
	QuantityChange int32  `json:"quantity_change" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
}

// UpdateStock updates the stock of a product by ID
// @Summary Update stock
// @Description Updates the stock of a product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param request body StockRequest true "Stock Details"
// @Success 200 {object} proto.UpdateStockResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/admin/products/{id}/stock [put]
func UpdateStock(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		var req proto.UpdateStockRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.Error("Failed to bind request", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: map[string]string{"format": err.Error()},
			})
			return
		}

		req.ProductId = productID

		resp, err := productClient.UpdateStock(context.Background(), &req)
		if err != nil {
			utils.Error("Failed to update stock", map[string]interface{}{
				"error":           err,
				"product_id":      productID,
				"quantity_change": req.QuantityChange,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to update stock",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to update stock", map[string]interface{}{
				"error":      resp.Message,
				"product_id": productID,
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

// GetInventoryLogs fetches inventory logs for a product
// @Summary Get inventory logs
// @Description Fetches inventory logs for a specific product
// @Tags Products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Product ID"
// @Param page query integer false "Page number"
// @Param limit query integer false "Number of items per page"
// @Param sort_by query string false "Sort by column"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param filter_column query string false "Filter column"
// @Param filter_operator query string false "Filter operator"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.GetInventoryLogsResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/admin/products/{id}/logs [get]
func GetInventoryLogs(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")
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

		resp, err := productClient.GetInventoryLogs(context.Background(), &proto.GetInventoryLogsRequest{
			ProductId: productID,
			Filter:    filter,
			SortBy:    sortBy,
			SortOrder: sortOrder,
			Page:      int32(page),
			Limit:     int32(limit),
		})

		if err != nil {
			utils.Error("Failed to get inventory logs", map[string]interface{}{
				"error":      err,
				"product_id": productID,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get inventory logs",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to get inventory logs", map[string]interface{}{
				"error": resp,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to get inventory logs",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
