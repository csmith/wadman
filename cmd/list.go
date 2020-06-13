package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(listCommand)
}

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "List currently installed addons",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		disabled, err := install.DisabledAddons()
		if err != nil {
			fmt.Printf("Unable to list disabled addons: %v\n", err)
			disabled = make(map[string]bool)
		}

		fmt.Printf("%d addons installed:\n\n", len(config.Addons))
		for i := range config.Addons {
			addon := config.Addons[i]
			count := 0
			for d := range addon.Directories {
				if disabled[addon.Directories[d]] {
					count++
				}
			}

			var status string
			if count == len(addon.Directories) {
				status = " (DISABLED)"
			} else if count > 0 {
				status = " (PARTIALLY DISABLED)"
			}

			fmt.Printf("[%6d] %s%s\n", addon.Id, addon.Name, status)
		}
	},
}
