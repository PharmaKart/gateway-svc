package grpc

import (
	"context"

	pb "github.com/PharmaKart/gateway-svc/internal/proto"
	"google.golang.org/grpc"
)

type AuthClient interface {
	Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error)
	Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
	VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error)
}

type authClient struct {
	client pb.AuthServiceClient
}

func NewAuthServiceClient(conn *grpc.ClientConn) AuthClient {
	return &authClient{
		client: pb.NewAuthServiceClient(conn),
	}
}

func (c *authClient) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return c.client.Register(ctx, req)
}

func (c *authClient) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return c.client.Login(ctx, req)
}

func (c *authClient) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	return c.client.VerifyToken(ctx, req)
}
