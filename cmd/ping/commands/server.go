//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//

// This file contains the specification of the API program in charge of launching the gRPC server.

package commands

import (
	"github.com/nalej/network-manager/internal/pkg/server"
	"github.com/spf13/cobra"
)


var portServer int

var apiCmd = &cobra.Command{
	Use:   "server",
	Short: "Launch the gRPC services",
	Long:  `Launch all registered gRPC services`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		s := server.NewServer(portServer)
		s.Launch()
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().IntVar(&portServer, "portServer", 8000, "Port to launch the gRPC server")
}