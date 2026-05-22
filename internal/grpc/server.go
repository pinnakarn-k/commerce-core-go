package grpc

import (
	"net"

	orderv1 "github.com/pinnakarn-k/commerce-core-go/internal/gen/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewServer(orderQueryHandler orderv1.OrderQueryServiceServer) *grpc.Server {
	server := grpc.NewServer()

	orderv1.RegisterOrderQueryServiceServer(
		server,
		orderQueryHandler,
	)

	reflection.Register(server)

	return server
}

func Run(server *grpc.Server, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return server.Serve(lis)
}
