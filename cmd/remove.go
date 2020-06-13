package cmd

import (
	"fmt"
	"github.com/csmith/wadman/internal"
	"github.com/spf13/cobra"
	"strconv"
)

func init() {
	rootCommand.AddCommand(removeCommand)
}

var removeCommand = &cobra.Command{
	Use:   "remove <id [id [id [...]]]>",
	Short: "Remove previously installed addons",
	Args:  addonIdArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer saveConfig()

		var newAddons []*internal.Addon
		for i := range config.Addons {
			addon := config.Addons[i]

			target := strconv.Itoa(addon.Id)
			found := false
			for n := range args {
				if args[n] == target {
					found = true
					break
				}
			}

			if found {
				if err := install.RemoveAddons(addon.Directories); err != nil {
					fmt.Printf("Failed to delete addon '%s': %v\n", addon.Name, err)
				} else {
					fmt.Printf("Removed addon '%s'\n", addon.Name)
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
