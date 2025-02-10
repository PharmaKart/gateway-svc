package handlers

import (
	"context"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/PharmaKart/gateway-svc/pkg/utils"

	// "github.com/PharmaKart/gateway-svc/pkg/s3"
	"github.com/gin-gonic/gin"
)

type ProductRequest struct {
	Name                 string  `json:"name" form:"name" binding:"required" example:"Paracetamol"`
	Description          string  `json:"description" form:"description" binding:"required" example:"Pain relief medication"`
	Price                float64 `json:"price" form:"price" binding:"required,gt=0" example:"9.99"`
	Stock                int32   `json:"stock" form:"stock" binding:"required,gte=0" example:"100"`
	RequiresPrescription bool    `json:"requires_prescription" form:"requires_prescription" example:"true"`
}

type Product struct {
	ProductRequest
	Image *multipart.FileHeader `form:"image" binding:"required" swaggerignore:"true"`
}

type SwaggerProduct struct {
	Name                 string  `json:"name" example:"Paracetamol"`
	Description          string  `json:"description" example:"Pain relief medication"`
	Price                float64 `json:"price" example:"9.99"`
	Stock                int32   `json:"stock" example:"100"`
	RequiresPrescription bool    `json:"requires_prescription" example:"true"`
	Image                string  `json:"image" format:"binary"`
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
// @Param image formData file true "Product Image"
// @Success 200 {object} proto.CreateProductResponse
// @Router /api/v1/products [post]
func CreateProduct(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Product
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate file type
		allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
		ext := filepath.Ext(req.Image.Filename)
		if !allowedExtensions[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only .jpg, .jpeg, and .png are allowed"})
			return
		}

		// Upload image to S3
		imageURL, err := utils.UploadImageToS3(c, "products", req.Image)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image to S3: " + err.Error()})
			return
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
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
// @Router /api/v1/products/{id} [get]
func GetProduct(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		resp, err := productClient.GetProduct(context.Background(), &proto.GetProductRequest{
			ProductId: productID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product: " + err.Error()})
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
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param filter query string false "Filter field"
// @Param filter_value query string false "Filter value"
// @Success 200 {object} proto.ListProductsResponse
// @Router /api/v1/products [get]
func GetProducts(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := utils.GetIntQueryParam(c, "page", 1)
		limit := utils.GetIntQueryParam(c, "limit", 10)
		sortBy := c.Query("sort_by")
		sortOrder := c.Query("sort_order")
		filter := c.Query("filter")
		filterValue := c.Query("filter_value")

		resp, err := productClient.ListProducts(context.Background(), &proto.ListProductsRequest{
			Page:        int32(page),
			Limit:       int32(limit),
			SortBy:      sortBy,
			SortOrder:   sortOrder,
			Filter:      filter,
			FilterValue: filterValue,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products: " + err.Error()})
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
// @Param request body Product true "Product Details"
// @Success 200 {object} proto.UpdateProductResponse
// @Router /api/v1/products/{id} [put]
func UpdateProduct(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		var req proto.Product
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := productClient.UpdateProduct(context.Background(), &proto.UpdateProductRequest{
			ProductId: productID,
			Product:   &req,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product: " + err.Error()})
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
// @Router /api/v1/products/{id} [delete]
func DeleteProduct(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		resp, err := productClient.DeleteProduct(context.Background(), &proto.DeleteProductRequest{
			ProductId: productID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

type StockRequest struct {
	Quantity int32  `json:"quantity" binding:"required,gte=0"`
	Reason   string `json:"reason" binding:"required"`
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
// @Router /api/v1/products/{id}/stock [put]
func UpdateStock(productClient grpc.ProductClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		var req proto.UpdateStockRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := productClient.UpdateStock(context.Background(), &proto.UpdateStockRequest{
			ProductId: productID,
			Quantity:  req.Quantity,
			Reason:    req.Reason,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
