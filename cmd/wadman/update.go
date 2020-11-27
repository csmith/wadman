package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func init() {
	rootCommand.AddCommand(updateCommand)
	updateCommand.Flags().BoolVarP(&force, "force", "f", false, "Replace all addons with the latest version")
	updateCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show debug information when checking for updates")
}

var force bool
var verbose bool

var updateCommand = &cobra.Command{
	Use:   "update [id [id ...]]",
	Short: "Update installed addons",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer saveConfig()

		matched := 0
		filtered := len(args) > 0
		included := toIdMap(args)

		var debug io.Writer
		if verbose {
			debug = os.Stdout
		} else {
			debug = ioutil.Discard
		}

		for i := range config.Addons {
			if !filtered || included[config.Addons[i].ShortName()] {
				matched++
				updated, version, err := config.Addons[i].Update(install, debug, force)
				if err != nil {
					fmt.Printf("Unable to update addon '%s': %v\n", config.Addons[i].DisplayName(), err)
				} else if force {
					fmt.Printf("Reinstalled addon '%s' at version %s\n", config.Addons[i].DisplayName(), version)
				} else if updated {
					fmt.Printf("Updated addon '%s' to version %s\n", config.Addons[i].DisplayName(), version)
				}
			}
		}

		if len(config.Addons) == 0 {
			fmt.Printf("No addons configured. Add addons to the config file: %s\n", configPath)
		} else {
			fmt.Printf("Finished checking %d addons\n", matched)
		}
	},
}

func toIdMap(args []string) map[string]bool {
	res := make(map[string]bool)
	for _, a := range args {
		if i, err := strconv.Atoi(a); err == nil {
			res[fmt.Sprintf("curse:%d", i)] = true
		}
		res[strings.ToLower(a)] = true
	}
	return res
}
