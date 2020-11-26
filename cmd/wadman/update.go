package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strconv"
)

func init() {
	rootCommand.AddCommand(updateCommand)
	updateCommand.Flags().BoolVarP(&force, "force", "f", false, "Replace all addons with the latest version")
	updateCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show debug information when checking for updates")
}

var force bool
var verbose bool

var updateCommand = &cobra.Command{
	Use:   "update [id [id ...]]",
	Short: "Update installed addons",
	Args:  optionalAddonIdArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer saveConfig()

		matched := 0
		filtered := len(args) > 0
		included := toIdMap(args)

		for i := range config.Addons {
			if !filtered || included[config.Addons[i].Id] {
				matched++
				if err := install.CheckUpdates(config.Addons[i], force, verbose); err != nil {
					fmt.Printf("Unable to update addon '%s': %v\n", config.Addons[i].Name, err)
				}
			}
		}

		if len(config.Addons) == 0 {
			log.Printf("No addons configured. Add addons to the config file: %s", configPath)
		} else {
			log.Printf("Finished checking %d addons", matched)
		}
	},
}

func toIdMap(args []string) map[int]bool {
	res := make(map[int]bool)
	for _, a := range args {
		i, _ := strconv.Atoi(a)
		res[i] = true
	}
	return res
}
