//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//

// This file contains the base command definition for the launching the application.

package commands

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

var debugLevel bool
var consoleLogging bool

var rootCmd = &cobra.Command{
	Use:   "grpcserver",
	Short: "gRPC example server",
	Long:  "gRPC example server for testing Go gRPC",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugLevel, "debug", false, "Set debug level")
	rootCmd.PersistentFlags().BoolVar(&consoleLogging, "consoleLogging", false, "Pretty print logging")
}

// Execute triggers the parsing and execution of the user specified command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// SetupLogging sets the debug level and console logging if required.
func SetupLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugLevel {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if consoleLogging {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
