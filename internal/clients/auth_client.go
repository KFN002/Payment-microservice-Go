package clients

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // Import insecure package

	pb "gitlab.crja72.ru/gospec/go8/payment/internal/payment-service/proto"
)

type AuthClient struct {
	client pb.AuthClient
	conn   *grpc.ClientConn
}

func NewAuthClient(grpcServerAddress string) (*AuthClient, error) {
	conn, err := grpc.NewClient(grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthClient(conn)
	return &AuthClient{
		client: client,
		conn:   conn,
	}, nil
}

func (a *AuthClient) Close() {
	if a.conn != nil {
		a.conn.Close()
	}
}

// GetUserById Получение данных пользователя по id
func (a *AuthClient) GetUserById(ctx context.Context, id string) (*pb.GetUserByIdResponse, error) {
	request := &pb.GetUserByIdRequest{
		Id: id,
	}
	return a.client.GetUserById(ctx, request)
}
