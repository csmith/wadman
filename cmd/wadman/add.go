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
			var addon wadman.Addon
			if strings.HasPrefix(args[i], "wowi:") {
				target, _ := strconv.Atoi(strings.TrimPrefix(args[i], "wowi:"))
				addon = wadman.NewWowInterfaceAddon(target)
			} else {
				// Assume it's CurseForge if it's unprefixed
				target, _ := strconv.Atoi(strings.TrimPrefix(args[i], "curse:"))
				addon = wadman.NewCurseForgeAddon(target)
			}

			if _, err := addon.Update(install, ioutil.Discard, false); err != nil {
				fmt.Printf("Unable to install addon %s: %v\n", args[i], err)
			} else {
				fmt.Printf("Installed addon '%s' version %s\n", addon.DisplayName(), addon.CurrentVersion())
				config.Addons = append(config.Addons, addon)
			}
		}
	},
}
