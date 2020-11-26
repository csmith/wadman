package main

import (
	"fmt"
	"github.com/csmith/wadman"
	"github.com/spf13/cobra"
	"strconv"
)

func init() {
	rootCommand.AddCommand(addCommand)
}

var addCommand = &cobra.Command{
	Use:   "add <id [id [id [...]]]>",
	Short: "Download and install new addons",
	Args:  requiredAddonIdArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer saveConfig()

		for i := range args {
			target, _ := strconv.Atoi(args[i])
			addon := &wadman.CurseForgeAddon{Id: target}
			if err := install.CheckUpdates(addon, false, false); err != nil {
				fmt.Printf("Unable to install addon #%d: %v\n", target, err)
			}
			config.Addons = append(config.Addons, addon)
		}
	},
}
