package server

import (
	"fmt"
	"github.com/nalej/networking/internal/pkg/services/ping"
	pbPing "github.com/nalej/example-grpc/ping"
	"github.com/rs/zerolog/log"
	"github.com/nalej/grpc-utils/pkg/tools"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server struct {
	Configuration Config
	Server * tools.GenericGRPCServer
}

func NewServer(config Config) *Server {
	return &Server{
		config,
		tools.NewGenericGRPCServer(uint32(config.Port)),
	}
}

func (s *Server) Launch() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	pingHandler := ping.NewHandler()

	grpcServer := grpc.NewServer()
	pbPing.RegisterPingServer(grpcServer, pingHandler)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}