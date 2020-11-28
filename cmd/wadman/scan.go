package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

func init() {
	rootCommand.AddCommand(scanCommand)
	scanCommand.Flags().BoolVar(&skipLoadOnDemand, "skip-load-on-demand",  true, "Skip load-on-demand addons")
}

var skipLoadOnDemand bool

var scanCommand = &cobra.Command{
	Use:   "scan",
	Short: "Scans for existing addons in the WoW install",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		addons, err := install.ListAddons()
		if err != nil {
			bail("Unable to read addon directory: %v", err)
		}

		addCmd := strings.Builder{}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Addon", "Scan result"})
		table.SetAutoWrapText(false)
		for i := range addons {
			metadata, _, err := install.ReadToc(addons[i])
			if err != nil {
				table.Append([]string{addons[i], fmt.Sprintf("error: %v", err)})
				continue
			}

			if (skipLoadOnDemand && metadata["loadondemand"] == "1") || metadata["x-part-of"] != "" {
				continue
			}

			status := strings.Builder{}
			if curse, ok := metadata["x-curse-project-id"]; ok {
				if _, err := strconv.Atoi(curse); err == nil{
					if status.Len() == 0 {
						addCmd.WriteString(" curse:")
						addCmd.WriteString(curse)
					}
					status.WriteString(" curse:")
					status.WriteString(curse)
				}
			}

			if wowi, ok := metadata["x-wowi-id"]; ok {
				if _, err := strconv.Atoi(wowi); err == nil {
					if status.Len() == 0 {
						addCmd.WriteString(" wowi:")
						addCmd.WriteString(wowi)
					}
					status.WriteString(" wowi:")
					status.WriteString(wowi)
				}
			}

			if status.Len() == 0 {
				status.WriteString("unknown")
			}

			table.Append([]string{addons[i], strings.TrimSpace(status.String())})
		}
		table.Render()
		if addCmd.Len() > 0 {
			fmt.Printf("\nTo add all addons, run:\n\n\twadman add%s\n", addCmd.String())
		}
	},
}
