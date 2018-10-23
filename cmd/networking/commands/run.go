//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//

// This file contains the specification of the API program in charge of launching the gRPC server.

package commands

import (
	"github.com/nalej/networking/internal/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = server.Config{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch the server API",
	Long:  `Launch the server API`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		log.Info().Msg("Launching API!")
		s := server.NewServer(config)
		s.Launch()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVar(&config.Port, "port", 8000, "Port to launch the gRPC server")
}
