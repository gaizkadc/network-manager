/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package commands

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/nalej/grpc-network-go"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// GRPC server address
var deleteEntryServer string
// Network ID
var deleteEntryOrganizationId string
// FQDN
var deleteEntryFqdn string

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
	deleteEntryCmd.Flags().StringVar(&deleteEntryOrganizationId, "orgid", "", "ID of the organization from which the DNS entry will be deleted")
	deleteEntryCmd.Flags().StringVar(&deleteEntryFqdn, "fqdn", "", "FQDN of the DNS entry")
	deleteEntryCmd.MarkFlagRequired("netid")
	deleteEntryCmd.MarkFlagRequired("fqdn")
}


func deleteEntry() {

	conn, err := grpc.Dial(deleteEntryServer, grpc.WithInsecure())

	if err!=nil{
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", deleteEntryServer)
	}

	client := grpc_network_go.NewDNSClient(conn)

	request := grpc_network_go.DeleteDNSEntryRequest{
		OrganizationId: deleteEntryOrganizationId,
		Fqdn:      deleteEntryFqdn,
	}

	deletedEntry, err := client.DeleteDNSEntry(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error deleting dns register %s", deleteEntryFqdn)
		return
	}

	log.Info().Msgf("%s", deletedEntry.String())
}