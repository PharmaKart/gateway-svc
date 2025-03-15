package grpc

import (
	"context"

	"github.com/PharmaKart/gateway-svc/internal/proto"
	"google.golang.org/grpc"
)

type OrderClient interface {
	PlaceOrder(ctx context.Context, req *proto.PlaceOrderRequest) (*proto.PlaceOrderResponse, error)
	GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error)
	ListCustomersOrders(ctx context.Context, req *proto.ListCustomersOrdersRequest) (*proto.ListCustomersOrdersResponse, error)
	ListAllOrders(ctx context.Context, req *proto.ListAllOrdersRequest) (*proto.ListAllOrdersResponse, error)
	UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusRequest) (*proto.UpdateOrderStatusResponse, error)
	GenerateNewPaymentUrl(ctx context.Context, req *proto.GenerateNewPaymentUrlRequest) (*proto.GenerateNewPaymentUrlResponse, error)
}

type orderClient struct {
	client proto.OrderServiceClient
}

func NewOrderServiceClient(conn *grpc.ClientConn) OrderClient {
	return &orderClient{
		client: proto.NewOrderServiceClient(conn),
	}
}

func (c *orderClient) PlaceOrder(ctx context.Context, req *proto.PlaceOrderRequest) (*proto.PlaceOrderResponse, error) {
	return c.client.PlaceOrder(ctx, req)
}

func (c *orderClient) GenerateNewPaymentUrl(ctx context.Context, req *proto.GenerateNewPaymentUrlRequest) (*proto.GenerateNewPaymentUrlResponse, error) {
	return c.client.GenerateNewPaymentUrl(ctx, req)
}

func (c *orderClient) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	return c.client.GetOrder(ctx, req)
}

func (c *orderClient) ListCustomersOrders(ctx context.Context, req *proto.ListCustomersOrdersRequest) (*proto.ListCustomersOrdersResponse, error) {
	return c.client.ListCustomersOrders(ctx, req)
}

func (c *orderClient) ListAllOrders(ctx context.Context, req *proto.ListAllOrdersRequest) (*proto.ListAllOrdersResponse, error) {
	return c.client.ListAllOrders(ctx, req)
}

func (c *orderClient) UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusRequest) (*proto.UpdateOrderStatusResponse, error) {
	return c.client.UpdateOrderStatus(ctx, req)
}
