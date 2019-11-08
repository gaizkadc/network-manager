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
var addNetworkServer string

// Application instance id
var appInstanceId string

// Organization ID
var addNetworkOrgId string

var addNetworkCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new network",
	Long:  `Add a new network`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		addNetwork()
	},
}

func init() {
	rootCmd.AddCommand(addNetworkCmd)
	addNetworkCmd.Flags().StringVar(&addNetworkServer, "server", "localhost:8000", "Networking manager server URL")
	addNetworkCmd.Flags().StringVar(&appInstanceId, "appinstanceid", "", "Application instance id")
	addNetworkCmd.Flags().StringVar(&addNetworkOrgId, "orgid", "", "Organization ID")
	addNetworkCmd.MarkFlagRequired("orgid")
	addNetworkCmd.MarkFlagRequired("appinstanceid")
}

func addNetwork() {

	conn, err := grpc.Dial(addNetworkServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", addNetworkServer)
	}

	client := grpc_network_go.NewNetworksClient(conn)

	request := grpc_network_go.AddNetworkRequest{
		Name:           appInstanceId,
		OrganizationId: addNetworkOrgId,
		AppInstanceId:  appInstanceId,
	}

	addedNetwork, err := client.AddNetwork(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msg("error adding network")
		return
	}

	log.Info().Msgf("%s", addedNetwork.String())
}
