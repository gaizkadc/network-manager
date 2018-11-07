package server

import (
	"fmt"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/network-manager/internal/pkg/consul"
	"github.com/nalej/network-manager/internal/pkg/server/dns"
	"github.com/nalej/network-manager/internal/pkg/server/networks"
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

	orgConn, err := grpc.Dial(s.Configuration.SystemModelURL, grpc.WithInsecure())
	if err != nil {
		log.Fatal().Msgf("impossible to establish connection with %s", s.Configuration.SystemModelURL)
		return
	}

	// Instantiate network manager
	netManager, err := networks.NewManager(orgConn, s.Configuration.ZTUrl, s.Configuration.ZTAccessToken)

	if err != nil {
		log.Fatal().Msg("failed creating network manager")
		return
	}

	netHandler := networks.NewHandler(*netManager)

	// Instantiate DNS manager
	consulClient, err := consul.NewConsulClient(s.Configuration.DNSUrl)
	if err != nil {
		log.Fatal().Msg("failed creating dns consul client")
		return
	}

	dnsManager, err := dns.NewManager(consulClient)
	if err != nil {
		log.Fatal().Msg("failed creating dns manager")
		return
	}

	dnsHandler := dns.NewHandler(*dnsManager)

	// gRPC Server
	grpcServer := grpc.NewServer()
	grpc_network_go.RegisterNetworksServer(grpcServer, netHandler)
	grpc_network_go.RegisterDNSServer(grpcServer, dnsHandler)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}