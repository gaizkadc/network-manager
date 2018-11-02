package commands

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/nalej/grpc-network-go"
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
}


func getNetwork() {

	conn, err := grpc.Dial(getNetworkServer, grpc.WithInsecure())

	if err!=nil{
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", getNetworkServer)
	}

	client := grpc_network_go.NewNetworksClient(conn)

	request := grpc_network_go.NetworkId{
		NetworkId: getNetworkId,
		OrganizationId: getNetworkOrgId,
	}

	retrievedNetwork, err := client.GetNetwork(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error retrieving network %s", getNetworkId)
		return
	}

	log.Info().Msgf("%s",retrievedNetwork.String())
}
