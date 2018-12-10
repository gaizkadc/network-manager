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
var deleteEntryServer string

// Network ID
var deleteEntryOrganizationId string

// ApplicationInstanceId
var deleteEntryAppInstanceId string

var deleteEntryCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a DNS entry",
	Long:  `Delete a DNS entry`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		deleteEntry()
	},
}

func init() {
	rootCmd.AddCommand(deleteEntryCmd)
	deleteEntryCmd.Flags().StringVar(&deleteEntryServer, "server", "localhost:8000", "Networking manager server URL")
	deleteEntryCmd.Flags().StringVar(&deleteEntryOrganizationId, "orgId", "", "ID of the organization from which the DNS entry will be deleted")
	deleteEntryCmd.Flags().StringVar(&deleteEntryAppInstanceId, "appId", "", "Application instance ID")
	deleteEntryCmd.MarkFlagRequired("netid")
	deleteEntryCmd.MarkFlagRequired("appId")
}

func deleteEntry() {

	conn, err := grpc.Dial(deleteEntryServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", deleteEntryServer)
	}

	client := grpc_network_go.NewDNSClient(conn)

	request := grpc_network_go.DeleteDNSEntryRequest{
		OrganizationId: deleteEntryOrganizationId,
		AppInstanceId:  deleteEntryAppInstanceId,
	}

	_, err = client.DeleteDNSEntry(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error deleting dns register for appId %s", deleteEntryAppInstanceId)
		return
	}

	log.Info().Msg("OK")
}
