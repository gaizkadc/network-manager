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
	runCmd.Flags().BoolVar(&config.UseTLS, "useTLS", true, "Use TLS to connect to the application cluster API")
	runCmd.Flags().StringVar(&config.CACertPath, "caCertPath", "", "Path for the CA certificate")
	runCmd.Flags().StringVar(&config.ClientCertPath, "clientCertPath", "", "Path for the client certificate")
	runCmd.Flags().BoolVar(&config.SkipServerCertValidation, "skipServerCertValidation", true, "Skip server cert validation")
}
