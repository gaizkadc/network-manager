package server

import (
	"fmt"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-utils/pkg/tools"
	"github.com/nalej/nalej-bus/pkg/bus/pulsar-comcast"
	"github.com/nalej/nalej-bus/pkg/queue/network/ops"
	"github.com/nalej/network-manager/internal/pkg/consul"
	"github.com/nalej/network-manager/internal/pkg/queue"
	"github.com/nalej/network-manager/internal/pkg/server/application"
	"github.com/nalej/network-manager/internal/pkg/server/dns"
	"github.com/nalej/network-manager/internal/pkg/server/networks"
	"github.com/nalej/network-manager/internal/pkg/server/servicedns"
	"github.com/nalej/network-manager/internal/pkg/utils"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Service struct {
	Configuration Config
	ConnHelper    *utils.ConnectionsHelper
	Server        *tools.GenericGRPCServer
}

func NewService(config Config) *Service {
	return &Service{
		Configuration: config,
		ConnHelper: utils.NewConnectionsHelper(config.UseTLS,config.ClientCertPath,config.CACertPath,config.SkipServerCertValidation),
		Server: tools.NewGenericGRPCServer(uint32(config.Port)),
	}
}

func (s *Service) Launch() {

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



	// Service Net application
	netAppManager := application.NewManager(smConn, s.ConnHelper)
	servNetAppHandler := application.NewHandler(*netAppManager)

	// Queue manager
	log.Info().Str("queueURL", s.Configuration.QueueAddress).Msg("instantiate message queue")
	pulsarclient := pulsar_comcast.NewClient(s.Configuration.QueueAddress, nil)

	log.Info().Msg("initialize networks ops manager...")
	networkOpsConfig := ops.NewConfigNetworksOpsConsumer(1, ops.ConsumableStructsNetworkOpsConsumer{
		AuthorizeMember: true, DisauthorizeMember: true, AddDNSEntry: true, DeleteDNSEntry: true,
		InboundServiceProxy: true, OutboundService: true, AddConnection: true, RemoveConnection: true,
		AuthorizeZTConnection: true,})

	networkOpsConsumer,err := ops.NewNetworkOpsConsumer(pulsarclient, "network-manager-network-ops", true, networkOpsConfig)
	if err!=nil{
		log.Panic().Err(err).Msg("impossible to initialize network ops manager")
	}
	networkOpsQueue := queue.NewNetworkOpsHandler(netManager, dnsManager, netAppManager, networkOpsConsumer)
	networkOpsQueue.Run()
	log.Info().Msg("initialize network ops manager done")

	// gRPC Service
	grpcServer := grpc.NewServer()
	grpc_network_go.RegisterNetworksServer(grpcServer, netHandler)
	grpc_network_go.RegisterDNSServer(grpcServer, dnsHandler)
	grpc_network_go.RegisterServiceDNSServer(grpcServer, servDNSHandler)
	grpc_network_go.RegisterApplicationNetworkServer(grpcServer, servNetAppHandler)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}
