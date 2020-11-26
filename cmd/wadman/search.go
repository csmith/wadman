package main

import (
	"fmt"
	"github.com/csmith/wadman/curse"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(searchCommand)
}

var searchCommand = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for available addons",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		results, err := curse.SearchAddons(args[0])
		if err != nil {
			bail("Unable to search addons: %v", err)
		}

		for i := range results {
			fmt.Printf("[%6d] %s\n", results[i].Id, results[i].Name)
		}
	},
}
