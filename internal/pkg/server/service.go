/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package server

import (
	"fmt"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-utils/pkg/tools"
	"github.com/nalej/nalej-bus/pkg/bus/pulsar-comcast"
	"github.com/nalej/nalej-bus/pkg/queue/application/events"
	"github.com/nalej/nalej-bus/pkg/queue/network/ops"
	"github.com/nalej/network-manager/internal/pkg/consul"
	"github.com/nalej/network-manager/internal/pkg/queue"
	"github.com/nalej/network-manager/internal/pkg/server/application"
	"github.com/nalej/network-manager/internal/pkg/server/dns"
	"github.com/nalej/network-manager/internal/pkg/server/networks"
	"github.com/nalej/network-manager/internal/pkg/server/servicedns"
	"github.com/nalej/network-manager/internal/pkg/utils"
	"github.com/nalej/network-manager/internal/pkg/zt"
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
		ConnHelper:    utils.NewConnectionsHelper(config.UseTLS, config.ClientCertPath, config.CACertPath, config.SkipServerCertValidation),
		Server:        tools.NewGenericGRPCServer(uint32(config.Port)),
	}
}

func (s *Service) Launch() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	smConn, err := grpc.Dial(s.Configuration.SystemModelURL, grpc.WithInsecure())
	if err != nil {
		log.Fatal().Str("SystemModelURL", s.Configuration.SystemModelURL).Msg("impossible to establish connection with system-model")
		return
	}

	// Create ZTClient
	ztClient, err := zt.NewZTClient(s.Configuration.ZTUrl, s.Configuration.ZTAccessToken)

	if err != nil {
		log.Error().Err(err).Str("ZTUrl", s.Configuration.ZTAccessToken).Msg("impossible to create network for url")
		return
	}

	// Instantiate network manager
	netManager, err := networks.NewManager(smConn, ztClient, s.ConnHelper)
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
	netAppManager, err := application.NewManager(smConn, s.ConnHelper, ztClient)
	if err != nil {
		log.Fatal().Msg("failed creating netapp manager")
		return
	}
	servNetAppHandler := application.NewHandler(*netAppManager)

	// Queue manager
	log.Info().Str("queueURL", s.Configuration.QueueAddress).Msg("instantiate message queue")
	pulsarclient := pulsar_comcast.NewClient(s.Configuration.QueueAddress, nil)

	log.Info().Msg("initialize networks ops manager...")
	networkOpsConfig := ops.NewConfigNetworksOpsConsumer(1, ops.ConsumableStructsNetworkOpsConsumer{
		AuthorizeMember: true, DisauthorizeMember: true, AddDNSEntry: true, DeleteDNSEntry: true,
		InboundServiceProxy: true, OutboundService: true, AddConnection: true, RemoveConnection: true,
		AuthorizeZTConnection: true, RegisterZTConnecion: true})

	networkOpsConsumer, err := ops.NewNetworkOpsConsumer(pulsarclient, "network-manager-network-ops", true, networkOpsConfig)
	if err != nil {
		log.Panic().Err(err).Msg("impossible to initialize network ops manager")
	}
	networkOpsQueue := queue.NewNetworkOpsHandler(netManager, dnsManager, netAppManager, networkOpsConsumer)
	networkOpsQueue.Run()
	log.Info().Msg("initialize network ops manager done")

	// application events consumer
	log.Info().Msg("initialize application events manager")
	appEventsConfig := events.NewConfigApplicationEventsConsumer(1, events.ConsumableStructsApplicationEventsConsumer{
		DeploymentServiceUpdateRequest: true,
	})
	appEventsConsumer, err := events.NewApplicationEventsConsumer(pulsarclient, "network-manager-application-events", true, appEventsConfig)
	if err != nil {
		log.Panic().Err(err).Msg("impossible to initialize application events manager")
	}
	appEventsQueue := queue.NewAppEventsHandler(netAppManager, appEventsConsumer)
	appEventsQueue.Run()
	log.Info().Msg("initialize application events manager done")

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
