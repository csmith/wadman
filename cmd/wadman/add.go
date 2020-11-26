package main

import (
	"fmt"
	"github.com/csmith/wadman"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

func init() {
	rootCommand.AddCommand(addCommand)
}

var addCommand = &cobra.Command{
	Use:   "add <id [id [id [...]]]>",
	Short: "Download and install new addons",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		defer saveConfig()

		for i := range args {
			// TODO: Support other addon types here
			target, _ := strconv.Atoi(strings.TrimPrefix(args[i], "curse:"))
			addon := &wadman.CurseForgeAddon{Id: target}
			if err := addon.Update(install, false, false); err != nil {
				fmt.Printf("Unable to install addon #%d: %v\n", target, err)
			}
			config.Addons = append(config.Addons, addon)
		}
	},
}
