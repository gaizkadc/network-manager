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
var authorizeMemberServer string

// Organization ID
var authorizeMemberOrganizationId string

// Network ID
var authorizeMemberNetworkId string

// Member ID
var authorizeMemberMemberId string

var authorizeMemberCmd = &cobra.Command{
	Use:   "authorize",
	Short: "Authorize a member to join a network",
	Long:  `Authorize a member to join a network`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		authorizeMember()
	},
}

func init() {
	rootCmd.AddCommand(authorizeMemberCmd)
	authorizeMemberCmd.Flags().StringVar(&authorizeMemberServer, "server", "localhost:8000", "Networking manager server URL")
	authorizeMemberCmd.Flags().StringVar(&authorizeMemberOrganizationId, "orgid", "", "Organization ID")
	authorizeMemberCmd.Flags().StringVar(&authorizeMemberNetworkId, "netid", "", "Network ID")
	authorizeMemberCmd.Flags().StringVar(&authorizeMemberMemberId, "memberid", "", "Member ID")
	authorizeMemberCmd.MarkFlagRequired("orgid")
	authorizeMemberCmd.MarkFlagRequired("netid")
	authorizeMemberCmd.MarkFlagRequired("memberid")
}

func authorizeMember() {

	conn, err := grpc.Dial(authorizeMemberServer, grpc.WithInsecure())

	if err != nil {
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", authorizeMemberServer)
	}

	client := grpc_network_go.NewNetworksClient(conn)

	request := grpc_network_go.AuthorizeMemberRequest{
		OrganizationId: authorizeMemberOrganizationId,
		NetworkId:      authorizeMemberNetworkId,
		MemberId:       authorizeMemberMemberId,
	}

	_, err = client.AuthorizeMember(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error authorizing member %s", authorizeMemberMemberId)
		return
	}

	log.Info().Msg("OK")
}
