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
var listEntriesServer string

// Organization ID
var listEntriesOrganizationId string

var listEntriesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DNS entries of an organization",
	Long:  `List all DNS entries of an organization`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		listEntries()
	},
}

func init() {
	rootCmd.AddCommand(listEntriesCmd)
	listEntriesCmd.Flags().StringVar(&listEntriesServer, "server", "localhost:8000", "Networking manager server URL")
	listEntriesCmd.Flags().StringVar(&listEntriesOrganizationId, "orgid", "", "Organization ID")
	listEntriesCmd.MarkFlagRequired("orgid")
}

func listEntries() {

	conn, err := grpc.Dial(listEntriesServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", listEntriesServer)
	}

	client := grpc_network_go.NewDNSClient(conn)

	request := grpc_organization_go.OrganizationId{
		OrganizationId: listEntriesOrganizationId,
	}

	listEntries, err := client.ListEntries(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error listing dns register %s", listEntriesOrganizationId)
		return
	}

	if listEntries == nil {
		log.Info().Msgf("%s", listEntries.String())
	} else {
		log.Info().Msg("No entries to list")
	}
}
