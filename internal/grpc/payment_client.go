package grpc

import (
	"context"

	"github.com/PharmaKart/gateway-svc/internal/proto"
	"google.golang.org/grpc"
)

type PaymentClient interface {
	StorePayment(ctx context.Context, req *proto.StorePaymentRequest) (*proto.StorePaymentResponse, error)
	RefundPayment(ctx context.Context, req *proto.RefundPaymentRequest) (*proto.RefundPaymentResponse, error)
	GetPaymentByTransactionID(ctx context.Context, req *proto.GetPaymentByTransactionIDRequest) (*proto.GetPaymentResponse, error)
	GetPayment(ctx context.Context, req *proto.GetPaymentRequest) (*proto.GetPaymentResponse, error)
	GetPaymentByOrderID(ctx context.Context, req *proto.GetPaymentByOrderIDRequest) (*proto.GetPaymentResponse, error)
}

type paymentClient struct {
	client proto.PaymentServiceClient
}

func NewPaymentServiceClient(conn *grpc.ClientConn) PaymentClient {
	return &paymentClient{
		client: proto.NewPaymentServiceClient(conn),
	}
}

func (c *paymentClient) StorePayment(ctx context.Context, req *proto.StorePaymentRequest) (*proto.StorePaymentResponse, error) {
	return c.client.StorePayment(ctx, req)
}

func (c *paymentClient) RefundPayment(ctx context.Context, req *proto.RefundPaymentRequest) (*proto.RefundPaymentResponse, error) {
	return c.client.RefundPayment(ctx, req)
}

func (c *paymentClient) GetPaymentByTransactionID(ctx context.Context, req *proto.GetPaymentByTransactionIDRequest) (*proto.GetPaymentResponse, error) {
	return c.client.GetPaymentByTransactionID(ctx, req)
}

func (c *paymentClient) GetPayment(ctx context.Context, req *proto.GetPaymentRequest) (*proto.GetPaymentResponse, error) {
	return c.client.GetPayment(ctx, req)
}

func (c *paymentClient) GetPaymentByOrderID(ctx context.Context, req *proto.GetPaymentByOrderIDRequest) (*proto.GetPaymentResponse, error) {
	return c.client.GetPaymentByOrderID(ctx, req)
}
