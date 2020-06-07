package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	force = flag.Bool("force", false, "Whether to force re-download of all addons")
	add = flag.Int("add", -1, "Project ID of an addon to download")
	delete = flag.Int("delete", -1, "Project ID of an addon to delete")
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
		if err := check(conf.InstallPath, addon); err != nil {
			log.Printf("Unable to install addon #%d: %v", *add, err)
		}
		conf.Addons = append(conf.Addons, addon)
	} else if *delete > 0 {
		var newAddons []*Addon
		for i := range conf.Addons {
			addon := conf.Addons[i]
			if addon.Id == *delete {
				if err := remove(conf.InstallPath, addon); err != nil {
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
	} else {
		for i := range conf.Addons {
			if err := check(conf.InstallPath, conf.Addons[i]); err != nil {
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

func check(path string, addon *Addon) error {
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
	} else if dirsMissing(path, addon.Directories) {
		log.Printf("'%s': missing directories, reinstalling version %s", addon.Name, latest.DisplayName)
	} else {
		return nil
	}

	// Remove all the existing directories associated with the addon
	if err := remove(path, addon); err != nil {
		return err
	}

	// Deploy the new version
	dirs, err := install(latest.Url, path)
	if err != nil {
		return err
	}

	// Update our metadata
	addon.FileId = latest.FileId
	addon.Directories = dirs
	return nil
}

func remove(path string, addon *Addon) error {
	for i := range addon.Directories {
		if err := os.RemoveAll(filepath.Join(path, addon.Directories[i])); err != nil {
			return err
		}
	}
	return nil
}

func dirsMissing(path string, directories []string) bool {
	for i := range directories {
		target := filepath.Join(path, directories[i])
		info, err := os.Stat(target)
		if err != nil || !info.IsDir() {
			return true
		}
	}
	return false
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

func install(url, path string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(bytes.NewReader(b), res.ContentLength)
	if err != nil {
		return nil, err
	}

	dirs := make(map[string]bool)

	for i := range reader.File {
		err := func(f *zip.File) error {
			parts := strings.Split(f.Name, "/")
			dirs[parts[0]] = true

			target := filepath.Join(path, f.Name)
			if f.FileInfo().IsDir() {
				return os.MkdirAll(target, os.FileMode(0755))
			} else {
				in, err := f.Open()
				if err != nil {
					return err
				}
				defer in.Close()

				_ = os.MkdirAll(filepath.Dir(target), os.FileMode(0755))
				out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					return err
				}
				defer out.Close()

				if _, err := io.Copy(out, in); err != nil {
					return err
				}

				return nil
			}
		}(reader.File[i])
		if err != nil {
			return nil, err
		}
	}

	var dirSlice []string
	for d := range dirs {
		dirSlice = append(dirSlice, d)
	}
	return dirSlice, nil
}
