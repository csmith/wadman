package main

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(migrateCommand)
}

var migrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Forces a config file migration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		saveConfig()
	},
}
