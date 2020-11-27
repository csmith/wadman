package main

import (
	"fmt"
	"github.com/csmith/wadman"
	"github.com/spf13/cobra"
	"io/ioutil"
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
			_, version, err := addon.Update(install, ioutil.Discard, false)
			if err != nil {
				fmt.Printf("Unable to install addon #%d: %v\n", target, err)
			} else {
				fmt.Printf("Installed addon '%s' version %s\n", addon.DisplayName(), version)
				config.Addons = append(config.Addons, addon)
			}
		}
	},
}
