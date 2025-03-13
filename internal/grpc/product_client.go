package grpc

import (
	"context"

	"github.com/PharmaKart/gateway-svc/internal/proto"
	"google.golang.org/grpc"
)

type ProductClient interface {
	CreateProduct(ctx context.Context, req *proto.CreateProductRequest) (*proto.CreateProductResponse, error)
	GetProduct(ctx context.Context, req *proto.GetProductRequest) (*proto.GetProductResponse, error)
	ListProducts(ctx context.Context, req *proto.ListProductsRequest) (*proto.ListProductsResponse, error)
	UpdateProduct(ctx context.Context, req *proto.UpdateProductRequest) (*proto.UpdateProductResponse, error)
	DeleteProduct(ctx context.Context, req *proto.DeleteProductRequest) (*proto.DeleteProductResponse, error)
	UpdateStock(ctx context.Context, req *proto.UpdateStockRequest) (*proto.UpdateStockResponse, error)
	GetInventoryLogs(ctx context.Context, req *proto.GetInventoryLogsRequest) (*proto.GetInventoryLogsResponse, error)
}

type productClient struct {
	client proto.ProductServiceClient
}

func NewProductServiceClient(conn *grpc.ClientConn) ProductClient {
	return &productClient{
		client: proto.NewProductServiceClient(conn),
	}
}

func (c *productClient) CreateProduct(ctx context.Context, req *proto.CreateProductRequest) (*proto.CreateProductResponse, error) {
	return c.client.CreateProduct(ctx, req)
}

func (c *productClient) GetProduct(ctx context.Context, req *proto.GetProductRequest) (*proto.GetProductResponse, error) {
	return c.client.GetProduct(ctx, req)
}

func (c *productClient) ListProducts(ctx context.Context, req *proto.ListProductsRequest) (*proto.ListProductsResponse, error) {
	return c.client.ListProducts(ctx, req)
}

func (c *productClient) UpdateProduct(ctx context.Context, req *proto.UpdateProductRequest) (*proto.UpdateProductResponse, error) {
	return c.client.UpdateProduct(ctx, req)
}

func (c *productClient) DeleteProduct(ctx context.Context, req *proto.DeleteProductRequest) (*proto.DeleteProductResponse, error) {
	return c.client.DeleteProduct(ctx, req)
}

func (c *productClient) UpdateStock(ctx context.Context, req *proto.UpdateStockRequest) (*proto.UpdateStockResponse, error) {
	return c.client.UpdateStock(ctx, req)
}

func (c *productClient) GetInventoryLogs(ctx context.Context, req *proto.GetInventoryLogsRequest) (*proto.GetInventoryLogsResponse, error) {
	return c.client.GetInventoryLogs(ctx, req)
}
