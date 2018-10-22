//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//

// This file contains the specification of the API program in charge of launching the gRPC server.

package commands

import (
	"github.com/nalej/example-grpc-server/internal/app/grpcserver"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(apiCmd)
}

var port int

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Launch the gRPC services",
	Long:  `Launch all registered gRPC services`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		server := grpcserver.NewServer(port)
		server.Launch()
	},
}

func init() {
	apiCmd.Flags().IntVar(&port, "port", 8000, "Port to launch the gRPC server")
}
