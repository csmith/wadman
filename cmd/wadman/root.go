package main

import (
	"fmt"
	"github.com/csmith/wadman"
	"github.com/csmith/wadman/wow"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strings"
)

var (
	rootCommand = &cobra.Command{
		Use:   "wadman",
		Short: "Wadman is a tool for managing World of Warcraft addons",
	}

	configPath string
	config     *wadman.Config
	install    *wow.Install
)

func init() {
	cobra.OnInitialize(loadConfig, createInstall)
}

func loadConfig() {
	var err error

	configPath, err = wadman.ConfigPath()
	if err != nil {
		bail("Unable to build config path: %v", err)
	}

	config, err = wadman.LoadConfig(configPath)
	if err != nil {
		bail("Unable to load config from %s: %v", configPath, err)
	}

	if config.InstallPath == "" {
		if path, ok := wow.GuessPath(); ok {
			fmt.Printf("Detected WoW install at %s\n", path)
			config.InstallPath = path
		} else {
			bail("Unable to find WoW install. Please edit the config file manually: %s", configPath)
		}
	}
}

func saveConfig() {
	sort.Slice(config.Addons, func(i, j int) bool {
		return strings.Compare(config.Addons[i].DisplayName(), config.Addons[j].DisplayName()) < 0
	})

	if err := wadman.SaveConfig(configPath, config); err != nil {
		bail("Unable to save config file to %s: %v", configPath, err)
	}
}

func createInstall() {
	install = wow.NewWowInstall(config.InstallPath)
}

func bail(format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("%s\n", format), args)
	os.Exit(1)
}
