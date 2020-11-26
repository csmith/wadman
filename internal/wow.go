package internal

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"github.com/csmith/wadman"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GuessWowPath() (string, bool) {
	paths := []string{
		"${PROGRAMW6432}\\World of Warcraft\\_retail_",
		"${PROGRAMFILES(X86)}\\World of Warcraft\\_retail_",
		"${HOME}/Games/world-of-warcraft/drive_c/Program Files (x86)/World of Warcraft/_retail_",
	}

	for _, p := range paths {
		expanded := os.ExpandEnv(p)
		if _, err := os.Stat(expanded); err == nil {
			return expanded, true
		}
	}

	return "", false
}

type WowInstall struct {
	path       string
	addonsPath string
}

func NewWowInstall(path string) *WowInstall {
	return &WowInstall{
		path:       path,
		addonsPath: filepath.Join(path, "Interface", "AddOns"),
	}
}

// ListAddons returns a list of addons currently installed in the WoW addons directory.
func (w *WowInstall) ListAddons() ([]string, error) {
	fs, err := ioutil.ReadDir(w.addonsPath)
	if err != nil {
		return nil, err
	}

	var folders []string
	for i := range fs {
		folders = append(folders, fs[i].Name())
	}

	return folders, nil
}

// RemoveAddons removes the specified addons from the WoW directory addons directory.
func (w *WowInstall) RemoveAddons(names []string) error {
	for i := range names {
		if err := os.RemoveAll(filepath.Join(w.addonsPath, names[i])); err != nil {
			return err
		}
	}
	return nil
}

// InstallAddon downloads a ZIP file from the given URL and deploys it to the WoW addons directory, returning a
// slice of top-level folder names that were created.
func (w *WowInstall) InstallAddon(url string) ([]string, error) {
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

			target := filepath.Join(w.addonsPath, f.Name)
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

// HasAddons returns true if all the given addons exist in the WoW addons directory.
func (w *WowInstall) HasAddons(names []string) bool {
	for i := range names {
		target := filepath.Join(w.addonsPath, names[i])
		info, err := os.Stat(target)
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}

// DisabledAddons returns a map of addons that are disabled in the WoW client.
func (w *WowInstall) DisabledAddons() (map[string]bool, error) {
	matches, err := filepath.Glob(filepath.Join(w.path, "WTF", "Account", "*", "*", "*", "AddOns.txt"))
	if err != nil {
		return nil, err
	}

	disabled := make(map[string]bool)
	for i := range matches {
		err := func(path string) error {
			file, err := os.Open(path)
			if err != nil {
				return err
			}

			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasSuffix(line, ": disabled") {
					name := strings.TrimSuffix(line, ": disabled")
					disabled[name] = true
				}
			}

			return nil
		}(matches[i])

		if err != nil {
			return nil, err
		}
	}

	return disabled, nil
}

// CheckUpdates checks for and applies updates for the given addon.
// The addon name will be updated to match the server-side name.
// If force is true, the addon will always be redeployed even if it appears to be up-to-date.
func (w *WowInstall) CheckUpdates(addon *wadman.CurseForgeAddon, force, verbose bool) error {
	if verbose {
		fmt.Println()
		fmt.Printf("================================================================================\n")
		fmt.Printf("Checking for updates to addon %d (%s)\n\n", addon.Id, addon.Name)
	}
	details, err := GetAddon(addon.Id)
	if err != nil {
		return err
	}

	addon.Name = details.Name

	latest := LatestFile(details, verbose)
	if latest == nil {
		return fmt.Errorf("no releases found for addon %d (%s)", addon.Id, addon.Name)
	}

	if force {
		fmt.Printf("'%s': force updating to version %s\n", addon.Name, latest.DisplayName)
	} else if addon.FileId == 0 {
		fmt.Printf("'%s': installing version %s\n", addon.Name, latest.DisplayName)
	} else if latest.FileId != addon.FileId {
		fmt.Printf("'%s': updating to version %s\n", addon.Name, latest.DisplayName)
	} else if !w.HasAddons(addon.Directories) {
		fmt.Printf("'%s': missing directories, reinstalling version %s\n", addon.Name, latest.DisplayName)
	} else {
		if verbose {
			fmt.Printf(
				"No update found for '%s'. Installed file ID: %d, latest file ID: %d (version: %s)\n",
				addon.Name,
				addon.FileId,
				latest.FileId,
				latest.DisplayName,
			)
		}

		return nil
	}

	// Remove all the existing directories associated with the addon
	if err := w.RemoveAddons(addon.Directories); err != nil {
		return err
	}

	// Deploy the new version
	dirs, err := w.InstallAddon(latest.Url)
	if err != nil {
		return err
	}

	// Update our metadata
	addon.FileId = latest.FileId
	addon.Directories = dirs
	return nil
}
