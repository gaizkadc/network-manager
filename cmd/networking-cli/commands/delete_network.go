package commands

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/nalej/grpc-network-go"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// GRPC server address
var delNetworkServer string
// Network name
var delNetworkId string
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
	delNetworkCmd.Flags().StringVar(&delNetworkId, "netid", "", "Network ID")
	delNetworkCmd.Flags().StringVar(&delNetworkOrgId, "orgid", "", "Organization ID")
	delNetworkCmd.MarkFlagRequired("netid")
	delNetworkCmd.MarkFlagRequired("orgid")
}


func delNetwork() {

	conn, err := grpc.Dial(delNetworkServer, grpc.WithInsecure())

	if err!=nil{
		log.Fatal().Err(err).Msgf("impossible to connect to server %s", delNetworkServer)
	}

	client := grpc_network_go.NewNetworksClient(conn)

	request := grpc_network_go.DeleteNetworkRequest{
		NetworkId:           delNetworkId,
		OrganizationId: delNetworkOrgId,
	}

	deletedNetwork, err := client.DeleteNetwork(context.Background(), &request)
	if err != nil {
		log.Error().Err(err).Msgf("error deleting network %s", delNetworkId)
		return
	}

	log.Info().Msgf("%s", deletedNetwork.String())
}
