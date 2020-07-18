package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCommand.AddCommand(updateCommand)
	updateCommand.Flags().BoolVarP(&force, "force", "f", false, "Replace all addons with the latest version")
	updateCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show debug information when checking for updates")
}

var force bool
var verbose bool

var updateCommand = &cobra.Command{
	Use:   "update",
	Short: "Update all installed addons",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer saveConfig()

		for i := range config.Addons {
			if err := install.CheckUpdates(config.Addons[i], force, verbose); err != nil {
				fmt.Printf("Unable to update addon '%s': %v\n", config.Addons[i].Name, err)
			}
		}

		if len(config.Addons) == 0 {
			log.Printf("No addons configured. Add addons to the config file: %s", configPath)
		} else {
			log.Printf("Finished checking %d addons", len(config.Addons))
		}
	},
}
