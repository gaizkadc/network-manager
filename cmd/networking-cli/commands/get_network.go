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
var getNetworkServer string

// Network ID
var getNetworkId string

// Organization ID
var getNetworkOrgId string

var getNetworkCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an existing network",
	Long:  `Get an existing network`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		getNetwork()
	},
}

func init() {
	rootCmd.AddCommand(getNetworkCmd)
	getNetworkCmd.Flags().StringVar(&getNetworkServer, "server", "localhost:8000", "Networking manager server URL")
	getNetworkCmd.Flags().StringVar(&getNetworkId, "netid", "", "Networking ID")
	getNetworkCmd.Flags().StringVar(&getNetworkOrgId, "orgid", "", "Organization ID")
	getNetworkCmd.MarkFlagRequired("netid")
	getNetworkCmd.MarkFlagRequired("orgid")
}

func getNetwork() {

	conn, err := grpc.Dial(getNetworkServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", getNetworkServer)
	}

	client := grpc_network_go.NewNetworksClient(conn)

	request := grpc_network_go.NetworkId{
		NetworkId:      getNetworkId,
		OrganizationId: getNetworkOrgId,
	}

	retrievedNetwork, err := client.GetNetwork(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error retrieving network %s", getNetworkId)
		return
	}

	log.Info().Msgf("%s", retrievedNetwork.String())
}
