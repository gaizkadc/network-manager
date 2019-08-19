//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//
// This file contains the specification of the API program in charge of launching the gRPC server.

package commands

import (
	"github.com/nalej/network-manager/internal/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

var config = server.Config{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch network manager",
	Long:  `Launch network manager`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		log.Info().Msg("Launching network manager!")

		if config.ZTAccessToken == "" {
			config.ZTAccessToken = os.Getenv("ZT_ACCESS_TOKEN")
		}

		config.Print()
		err := config.Validate()
		if err != nil {
			log.Fatal().Err(err)
		}


		s := server.NewService(config)
		s.Launch()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVar(&config.Port, "port", 8000, "Port to launch the gRPC server")
	runCmd.Flags().StringVar(&config.SystemModelURL, "sm", "localhost:8800", "System Model URL")
	runCmd.Flags().StringVar(&config.ZTUrl, "zturl", "http://localhost:9993", "ZT Controller URL")
	runCmd.Flags().StringVar(&config.ZTAccessToken, "ztaccesstoken", "", "ZT Access Token")
	runCmd.Flags().StringVar(&config.DNSUrl, "dnsurl", "192.168.99.100:30500", "Consul DNS URL")
	runCmd.Flags().StringVar(&config.QueueAddress, "queueAddress", "localhost:6650", "Message queue (localhost:6650)")
	runCmd.Flags().BoolVar(&config.UseTLS,"useTLS", true, "Use TLS to connect to the application cluster API")
	runCmd.Flags().StringVar(&config.CaCertPath,"caCertPath", "", "Part for the CA certificate")
	runCmd.Flags().BoolVar(&config.SkipServerCertValidation, "skipServerCertValidation", true, "Skip CA authentication validation")
}
