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
			} else if strings.HasPrefix(args[i], "curse:") {
				target, _ := strconv.Atoi(strings.TrimPrefix(args[i], "curse:"))
				addon = wadman.NewCurseForgeAddon(target)
			} else {
				fmt.Printf("%s: Unrecognised addon type. Did you mean curse:%[1]s or wowi:%[1]s?\n", args[i])
				continue
			}

			if addonExists(addon.ShortName()) {
				fmt.Printf("%s: Addon is already installed.\n", args[i])
				continue
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

func addonExists(shortName string) bool {
	for i := range config.Addons {
		if config.Addons[i].ShortName() == shortName {
			return true
		}
	}
	return false
}
