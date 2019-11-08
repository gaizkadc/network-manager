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
