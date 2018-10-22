package grpcserver

import (
	"fmt"
	"github.com/nalej/example-grpc-server/internal/app/grpcserver/services/ping"
	"github.com/nalej/example-grpc-server/internal/app/grpcserver/services/tester"
	pbPing "github.com/nalej/example-grpc/ping"
	pbTester "github.com/nalej/example-grpc/tester"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server struct {
	Port int
}

func NewServer(port int) *Server {
	return &Server{port}
}

func (s *Server) Launch() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	pingHandler := ping.NewHandler()
	testerHandler := tester.NewHandler()

	grpcServer := grpc.NewServer()
	pbPing.RegisterPingServer(grpcServer, pingHandler)
	pbTester.RegisterTesterServer(grpcServer, testerHandler)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}
