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

var config = cfg.Config{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch the server API",
	Long:  `Launch the server API`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		log.Info().Msg("Launching API!")
		s := server.NewService(config)
		s.Run()
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().IntVar(&portServer, "portServer", 8000, "Port to launch the gRPC server")
}
