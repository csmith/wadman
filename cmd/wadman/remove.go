package main

import (
	"fmt"
	"github.com/csmith/wadman"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(removeCommand)
}

var removeCommand = &cobra.Command{
	Use:   "remove <id [id [id [...]]]>",
	Short: "Remove previously installed addons",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		defer saveConfig()

		included := toIdMap(args)

		var newAddons []wadman.Addon
		for i := range config.Addons {
			addon := config.Addons[i]

			if included[addon.ShortName()] {
				if err := install.RemoveAddons(addon.Dirs()); err != nil {
					fmt.Printf("Failed to delete addon '%s': %v\n", addon.DisplayName(), err)
				} else {
					fmt.Printf("Removed addon '%s'\n", addon.DisplayName())
				}
			} else {
				newAddons = append(newAddons, addon)
			}
		}

		if len(newAddons) == len(config.Addons) {
			fmt.Printf("No matching addons found\n")
		} else {
			config.Addons = newAddons
		}
	},
}
