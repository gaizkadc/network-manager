package server

import (
	"fmt"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-utils/pkg/tools"
	"github.com/nalej/nalej-bus/pkg/bus/pulsar-comcast"
	"github.com/nalej/nalej-bus/pkg/queue/network/ops"
	"github.com/nalej/network-manager/internal/pkg/consul"
	"github.com/nalej/network-manager/internal/pkg/queue"
	"github.com/nalej/network-manager/internal/pkg/server/dns"
	"github.com/nalej/network-manager/internal/pkg/server/networks"
	"github.com/nalej/network-manager/internal/pkg/server/servicedns"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server struct {
	Configuration Config
	Server        *tools.GenericGRPCServer
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

	smConn, err := grpc.Dial(s.Configuration.SystemModelURL, grpc.WithInsecure())
	if err != nil {
		log.Fatal().Msgf("impossible to establish connection with %s", s.Configuration.SystemModelURL)
		return
	}

	// Instantiate network manager
	netManager, err := networks.NewManager(smConn, s.Configuration.ZTUrl, s.Configuration.ZTAccessToken)

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

	// ServiceDNS
	servDNSManager := servicedns.NewManager(consulClient)
	servDNSHandler := servicedns.NewHandler(servDNSManager)

	// Queue manager
	log.Info().Str("queueURL", s.Configuration.QueueAddress).Msg("instantiate message queue")
	pulsarclient := pulsar_comcast.NewClient(s.Configuration.QueueAddress)

	log.Info().Msg("initialize networks ops manager...")
	networkOpsConfig := ops.NewConfigNetworksOpsConsumer(1, ops.ConsumableStructsNetworkOpsConsumer{
		AuthorizeMember: true, DisauthorizeMember: true})

	networkOpsConsumer,err := ops.NewNetworkOpsConsumer(pulsarclient, "network-manager-network-ops", true, networkOpsConfig)
	if err!=nil{
		log.Panic().Err(err).Msg("impossible to initialize network ops manager")
	}
	networkOpsQueue := queue.NewNetworkOpsHandler(netManager, networkOpsConsumer)
	networkOpsQueue.Run()
	log.Info().Msg("initialize network ops manager done")

	// gRPC Server
	grpcServer := grpc.NewServer()
	grpc_network_go.RegisterNetworksServer(grpcServer, netHandler)
	grpc_network_go.RegisterDNSServer(grpcServer, dnsHandler)
	grpc_network_go.RegisterServiceDNSServer(grpcServer, servDNSHandler)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}
