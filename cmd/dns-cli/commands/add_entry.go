/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package commands

import (
	"context"
	"github.com/nalej/grpc-network-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// GRPC server address
var addEntryServer string

// Organization ID
var addEntryOrganizationId string

// FQDN
var addEntryFqdn string

// IP
var addEntryIp string

// Service name
var addEntryServiceName string

var addEntryCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new DNS entry",
	Long:  `Add a new DNS entry`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		addEntry()
	},
}

func init() {
	rootCmd.AddCommand(addEntryCmd)
	addEntryCmd.Flags().StringVar(&addEntryServer, "server", "localhost:8000", "Networking manager server URL")
	addEntryCmd.Flags().StringVar(&addEntryOrganizationId, "orgId", "", "Organization ID")
	addEntryCmd.Flags().StringVar(&addEntryFqdn, "fqdn", "", "FQDN of the DNS entry")
	addEntryCmd.Flags().StringVar(&addEntryIp, "ip", "", "IP of the DNS entry")
	addEntryCmd.Flags().StringVar(&addEntryServiceName, "serviceName", "", "service name")
	addEntryCmd.MarkFlagRequired("orgId")
	addEntryCmd.MarkFlagRequired("fqdn")
	addEntryCmd.MarkFlagRequired("ip")
	addEntryCmd.MarkFlagRequired("serviceName")
}

func addEntry() {

	conn, err := grpc.Dial(addEntryServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", addEntryServer)
	}

	client := grpc_network_go.NewDNSClient(conn)

	request := grpc_network_go.AddDNSEntryRequest{
		OrganizationId: addEntryOrganizationId,
		ServiceName:    addEntryServiceName,
		Fqdn:           addEntryFqdn,
		Ip:             addEntryIp,
	}

	_, err = client.AddDNSEntry(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error adding dns register %s", addEntryServiceName)
		return
	}

	log.Info().Msg("OK")
}
