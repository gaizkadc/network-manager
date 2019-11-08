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

// ServiceName
var deleteEntryServiceName string

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
	deleteEntryCmd.Flags().StringVar(&deleteEntryServiceName, "serviceName", "", "Service name")
	deleteEntryCmd.MarkFlagRequired("orgId")
	deleteEntryCmd.MarkFlagRequired("serviceName")
}

func deleteEntry() {

	conn, err := grpc.Dial(deleteEntryServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", deleteEntryServer)
	}

	client := grpc_network_go.NewDNSClient(conn)

	request := grpc_network_go.DeleteDNSEntryRequest{
		OrganizationId: deleteEntryOrganizationId,
		ServiceName:    deleteEntryServiceName,
	}

	_, err = client.DeleteDNSEntry(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error deleting dns register for appId %s", deleteEntryServiceName)
		return
	}

	log.Info().Msg("OK")
}
