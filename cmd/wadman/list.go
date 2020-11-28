package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
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
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "ID", "Status"})
		table.SetAutoWrapText(false)
		for i := range config.Addons {
			addon := config.Addons[i]
			count := 0
			dirs := addon.Dirs()
			for d := range dirs {
				if disabled[dirs[d]] {
					count++
				}
			}

			var status string
			if count == len(dirs) {
				status = "disabled"
			} else if count > 0 {
				status = fmt.Sprintf("disabled (%d/%d)", count, len(dirs))
			}

			table.Append([]string{addon.DisplayName(), addon.ShortName(), status})
		}
		table.Render()
	},
}
