//
// Copyright (C) 2018 Daisho Group - All Rights Reserved
//

// This file defines the version command. While not necessarily required, it provides a base
// template to have several cobra commands in a project.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version of the current binary.
const Version = "0.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gRPC",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("gRPC Example server %s", Version))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
