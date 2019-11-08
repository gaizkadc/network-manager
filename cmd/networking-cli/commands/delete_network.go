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
var delNetworkServer string

// Network name
var delAppInstanceId string

// Organization ID
var delNetworkOrgId string

var delNetworkCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an existing network",
	Long:  `Delete an existing network`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		delNetwork()
	},
}

func init() {
	rootCmd.AddCommand(delNetworkCmd)
	delNetworkCmd.Flags().StringVar(&delNetworkServer, "server", "localhost:8000", "Networking manager server URL")
	delNetworkCmd.Flags().StringVar(&delAppInstanceId, "appinstanceid", "", "Application instance id")
	delNetworkCmd.Flags().StringVar(&delNetworkOrgId, "orgid", "", "Organization ID")
	delNetworkCmd.MarkFlagRequired("appinstanceid")
	delNetworkCmd.MarkFlagRequired("orgid")
}

func delNetwork() {

	conn, err := grpc.Dial(delNetworkServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", delNetworkServer)
	}

	client := grpc_network_go.NewNetworksClient(conn)

	request := grpc_network_go.DeleteNetworkRequest{
		AppInstanceId:  appInstanceId,
		OrganizationId: delNetworkOrgId,
	}

	_, err = client.DeleteNetwork(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error deleting network %s", delAppInstanceId)
		return
	}

	log.Info().Msg("OK")
}
