/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package commands

import (
	"context"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// GRPC server address
var listNetworksServer string

// Organization ID
var listNetworksOrgId string

var listNetworksCmd = &cobra.Command{
	Use:   "list",
	Short: "List existing networks",
	Long:  `List existing networks of an organization`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		listNetworks()
	},
}

func init() {
	rootCmd.AddCommand(listNetworksCmd)
	listNetworksCmd.Flags().StringVar(&listNetworksServer, "server", "localhost:8000", "Networking manager server URL")
	listNetworksCmd.Flags().StringVar(&listNetworksOrgId, "orgid", "", "Organization ID")
	listNetworksCmd.MarkFlagRequired("orgid")
}

func listNetworks() {

	conn, err := grpc.Dial(listNetworksServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", listNetworksServer)
	}

	client := grpc_network_go.NewNetworksClient(conn)

	request := grpc_organization_go.OrganizationId{
		OrganizationId: listNetworksOrgId,
	}

	retrievedNetworkList, err := client.ListNetworks(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error retrieving Organization %s", listNetworksOrgId)
		return
	}

	if retrievedNetworkList == nil {
		log.Info().Msgf("%s", retrievedNetworkList.String())
	} else {
		log.Info().Msg("No networks to list")
	}
}
