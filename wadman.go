package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

var (
	force = flag.Bool("force", false, "Whether to force re-download of all addons")
	add = flag.Int("add", -1, "Project ID of an addon to download")
	delete = flag.Int("delete", -1, "Project ID of an addon to delete")
	list = flag.Bool("list", false, "List the available addons")

	wow *WowInstall
)

func main() {
	flag.Parse()

	path, err := ConfigPath()
	if err != nil {
		log.Panicf("Unable to build config path: %v", err)
	}

	conf, err := LoadConfig(path)
	if err != nil {
		log.Panicf("Unable to load config from %s: %v", path, err)
	}

	wow = NewWowInstall(conf.InstallPath)

	defer func() {
		sort.Slice(conf.Addons, func(i, j int) bool {
			return strings.Compare(conf.Addons[i].Name, conf.Addons[j].Name) < 0
		})

		if err := SaveConfig(path, conf); err != nil {
			log.Panicf("Unable to save config file to %s: %v", path, err)
		}
	}()

	if *add > 0 {
		addon := &Addon{Id: *add}
		if err := check(addon); err != nil {
			log.Printf("Unable to install addon #%d: %v", *add, err)
		}
		conf.Addons = append(conf.Addons, addon)
	} else if *delete > 0 {
		var newAddons []*Addon
		for i := range conf.Addons {
			addon := conf.Addons[i]
			if addon.Id == *delete {
				if err := wow.RemoveAddons(addon.Directories); err != nil {
					log.Printf("Failed to delete addon: %v", err)
					return
				}
				log.Printf("Removed addon '%s'", addon.Name)
			} else {
				newAddons = append(newAddons, addon)
			}
		}

		if len(newAddons) == len(conf.Addons) {
			log.Printf("Addon not found: %d", *delete)
		} else {
			conf.Addons = newAddons
		}
	} else if *list {
		disabled, err := wow.DisabledAddons()
		if err != nil {
			log.Printf("Unable to list disabled addons: %v", err)
			disabled = make(map[string]bool)
		}

		fmt.Printf("%d addons installed:\n\n", len(conf.Addons))
		for i := range conf.Addons {
			addon := conf.Addons[i]
			count := 0
			for d := range addon.Directories {
				if disabled[addon.Directories[d]] {
					count++
				}
			}

			var status string
			if count == len(addon.Directories) {
				status = " (DISABLED)"
			} else if count > 0 {
				status = " (PARTIALLY DISABLED)"
			}

			fmt.Printf("[%6d] %s%s\n", addon.Id, addon.Name, status)
		}
	} else {
		for i := range conf.Addons {
			if err := check(conf.Addons[i]); err != nil {
				log.Printf("Unable to update addon #%d: %v", i, err)
			}
		}

		if len(conf.Addons) == 0 {
			log.Printf("No addons configured. Add addons to the config file: %s", path)
		} else {
			log.Printf("Finished checking %d addons", len(conf.Addons))
		}
	}
}

func check(addon *Addon) error {
	details, err := GetAddon(addon.Id)
	if err != nil {
		return err
	}

	addon.Name = details.Name

	latest := latestFile(details)
	if latest == nil {
		return fmt.Errorf("no releases found for addon %d (%s)", addon.Id, addon.Name)
	}

	if *force {
		log.Printf("'%s': force updating to version %s", addon.Name, latest.DisplayName)
	} else if addon.FileId == 0 {
		log.Printf("'%s': installing version %s", addon.Name, latest.DisplayName)
	} else if latest.FileId != addon.FileId {
		log.Printf("'%s': updating to version %s", addon.Name, latest.DisplayName)
	} else if !wow.HasAddons(addon.Directories) {
		log.Printf("'%s': missing directories, reinstalling version %s", addon.Name, latest.DisplayName)
	} else {
		return nil
	}

	// Remove all the existing directories associated with the addon
	if err := wow.RemoveAddons(addon.Directories); err != nil {
		return err
	}

	// Deploy the new version
	dirs, err := wow.InstallAddon(latest.Url)
	if err != nil {
		return err
	}

	// Update our metadata
	addon.FileId = latest.FileId
	addon.Directories = dirs
	return nil
}

func latestFile(details *AddonResponse) *AddonFile {
	var (
		latestTime time.Time
		latestFile *AddonFile
	)

	for i := range details.Files {
		f := details.Files[i]
		if f.Flavour == "wow_retail" && f.Type <= Beta && !f.Alternate {
			if f.Date.After(latestTime) {
				latestTime = f.Date
				latestFile = &f
			}
		}
	}

	return latestFile
}
