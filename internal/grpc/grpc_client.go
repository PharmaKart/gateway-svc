package grpc

import (
	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient interface {
	Conn() *grpc.ClientConn
	Close()
}

type grpcClient struct {
	conn *grpc.ClientConn
}

func NewClient(url string) (GrpcClient, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &grpcClient{conn: conn}, nil
}

func (c *grpcClient) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *grpcClient) Close() {
	if err := c.conn.Close(); err != nil {
		utils.Logger.Error("Failed to close gRPC connection", map[string]interface{}{
			"error": err,
		})
	}
}
