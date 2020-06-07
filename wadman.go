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

	for i := range conf.Addons {
		if err := check(conf.InstallPath, conf.Addons[i]); err != nil {
			log.Printf("Unable to update addon #%d: %v\n", i, err)
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
	} else if latest.FileId != addon.FileId {
		log.Printf("'%s': updating to version %s", addon.Name, latest.DisplayName)
	} else if dirsMissing(path, addon.Directories) {
		log.Printf("'%s': missing directories, reinstalling version %s", addon.Name, latest.DisplayName)
	} else {
		return nil
	}

	// Remove all the existing directories associated with the addon
	for i := range addon.Directories {
		if err := os.RemoveAll(filepath.Join(path, addon.Directories[i])); err != nil {
			return err
		}
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
