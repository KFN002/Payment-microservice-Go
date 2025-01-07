package clients

import (
	"context"
	"errors"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/stretchr/testify/assert"
	"gitlab.crja72.ru/gospec/go8/payment/internal/payment-service/proto"
)

const bufSize = 1024 * 1024

var listener *bufconn.Listener

func init() {
	listener = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	proto.RegisterAuthServer(grpcServer, &MockAuthServer{})
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			panic(err)
		}
	}()
}

type MockAuthServer struct {
	proto.UnimplementedAuthServer
}

func (m *MockAuthServer) GetUserById(ctx context.Context, req *proto.GetUserByIdRequest) (*proto.GetUserByIdResponse, error) {
	if req.Id == "valid-id" {
		return &proto.GetUserByIdResponse{YoomoneyId: req.Id, Name: "Test User"}, nil
	}
	return nil, errors.New("user not found")
}

func bufDialer(ctx context.Context, s string) (net.Conn, error) {
	return listener.Dial()
}

func TestNewAuthClient(t *testing.T) {
	client, err := NewAuthClient("")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	defer client.Close()
}

func TestAuthClient_GetUserById_ValidID(t *testing.T) {
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := proto.NewAuthClient(conn)
	authClient := &AuthClient{
		client: client,
		conn:   conn,
	}

	response, err := authClient.GetUserById(context.Background(), "valid-id")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "valid-id", response.YoomoneyId)
	assert.Equal(t, "Test User", response.Name)
}

func TestAuthClient_GetUserById_InvalidID(t *testing.T) {
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := proto.NewAuthClient(conn)
	authClient := &AuthClient{
		client: client,
		conn:   conn,
	}

	response, err := authClient.GetUserById(context.Background(), "invalid-id")
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "user not found")
}
