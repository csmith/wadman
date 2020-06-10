package main

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

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
