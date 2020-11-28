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
		table.SetHeader([]string{"ID", "Name", "Version", "Last updated", "Status"})
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

			var lastUpdated string
			if addon.LastUpdated().IsZero() {
				lastUpdated = ""
			} else {
				lastUpdated = addon.LastUpdated().Format("2006-01-02 15:04")
			}


			table.Append([]string{addon.ShortName(), addon.DisplayName(), addon.CurrentVersion(), lastUpdated, status})
		}
		table.Render()
	},
}
