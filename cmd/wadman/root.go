package main

import (
	"fmt"
	"github.com/csmith/wadman/internal"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	rootCommand = &cobra.Command{
		Use:   "wadman",
		Short: "Wadman is a tool for managing World of Warcraft addons",
	}

	configPath string
	config     *internal.Config
	install    *internal.WowInstall
)

func init() {
	cobra.OnInitialize(loadConfig, createInstall)
}

func loadConfig() {
	var err error

	configPath, err = internal.ConfigPath()
	if err != nil {
		bail("Unable to build config path: %v", err)
	}

	config, err = internal.LoadConfig(configPath)
	if err != nil {
		bail("Unable to load config from %s: %v", configPath, err)
	}

	if config.InstallPath == "" {
		if path, ok := internal.GuessWowPath(); ok {
			fmt.Printf("Detected WoW install at %s\n", path)
			config.InstallPath = path
		} else {
			bail("Unable to find WoW install. Please edit the config file manually: %s", configPath)
		}
	}
}

func saveConfig() {
	sort.Slice(config.Addons, func(i, j int) bool {
		return strings.Compare(config.Addons[i].Name, config.Addons[j].Name) < 0
	})

	if err := internal.SaveConfig(configPath, config); err != nil {
		bail("Unable to save config file to %s: %v", configPath, err)
	}
}

func createInstall() {
	install = internal.NewWowInstall(config.InstallPath)
}

func requiredAddonIdArgs(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("requires at least one ID")
	}

	for i := range args {
		if _, err := strconv.Atoi(args[i]); err != nil {
			return fmt.Errorf("invalid ID at argument %d: %v", i + 1, err)
		}
	}

	return nil
}

func optionalAddonIdArgs(_ *cobra.Command, args []string) error {
	for i := range args {
		if _, err := strconv.Atoi(args[i]); err != nil {
			return fmt.Errorf("invalid ID at argument %d: %v", i + 1, err)
		}
	}

	return nil
}

func bail(format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("%s\n", format), args)
	os.Exit(1)
}
