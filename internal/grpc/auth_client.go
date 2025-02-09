package grpc

import (
	"context"

	"github.com/PharmaKart/gateway-svc/internal/proto"
	"google.golang.org/grpc"
)

type AuthClient interface {
	Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error)
	Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error)
	VerifyToken(ctx context.Context, req *proto.VerifyTokenRequest) (*proto.VerifyTokenResponse, error)
}

type authClient struct {
	client proto.AuthServiceClient
}

func NewAuthServiceClient(conn *grpc.ClientConn) AuthClient {
	return &authClient{
		client: proto.NewAuthServiceClient(conn),
	}
}

func (c *authClient) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	return c.client.Register(ctx, req)
}

func (c *authClient) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	return c.client.Login(ctx, req)
}

func (c *authClient) VerifyToken(ctx context.Context, req *proto.VerifyTokenRequest) (*proto.VerifyTokenResponse, error) {
	return c.client.VerifyToken(ctx, req)
}
