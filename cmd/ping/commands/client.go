/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

// This is an example of an executable command.

package commands

import (
	"github.com/nalej/networking/pkg/pinger"
	"github.com/spf13/cobra"
)

var numPings int
var targetPort int
var targetHost string

var pingCmd = &cobra.Command{
	Use:   "client",
	Short: "Send a ping to the gRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		a := pinger.NewPinger(targetHost, targetPort)
		a.Ping(numPings)
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	pingCmd.Flags().IntVar(&numPings, "numPings", 5, "Number of messages to send")
	pingCmd.Flags().IntVar(&targetPort, "port", 3000, "Port")
	pingCmd.Flags().StringVar(&targetHost, "host", "localhost", "Target host")
}
