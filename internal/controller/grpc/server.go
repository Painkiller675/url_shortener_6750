package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Painkiller675/url_shortener_6750/internal/controller"
	"github.com/Painkiller675/url_shortener_6750/internal/protos"
	"google.golang.org/grpc"
)

// Server - the router(like REST controller here) and the thingy that provides us with some handlers
type Server struct {
	protos.UnimplementedShortenerServer // it calls that we use gRPC here
	business                            controller.Business
}

func (s *Server) PingDB(ctx context.Context, _ *protos.PingDBRequest) (*protos.PingDBResponse, error) {
	if err := s.business.PingDB(ctx); err != nil {
		return nil, fmt.Errorf("err: %w", err)
	}
	return &protos.PingDBResponse{}, nil
}

// Serve - relates grpcServer with Shorten service  and launches the gRPC server
func Serve(srv *Server) error {
	lis, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		return fmt.Errorf("failed to run gRPC server: %v", err)
	}
	grpcServer := grpc.NewServer()
	protos.RegisterShortenerServer(grpcServer, srv)

	log.Println("gRPC server running")

	return grpcServer.Serve(lis)
}
