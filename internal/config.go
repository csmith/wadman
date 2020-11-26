package internal

import (
	"encoding/json"
	"fmt"
	"github.com/csmith/wadman"
	"os"
	"path/filepath"
)

// configVersion specifies the maximum version of the config file this build of wadman supports
// version 1 was the original config format
// version 2 changed the install_path field to the base _retail_ directory instead of the addons directory
// version 3 added a type field to addons
const configVersion = 3

type Config struct {
	InstallPath string
	Addons      []*wadman.CurseForgeAddon
}

func ConfigPath() (string, error) {
	basePath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(basePath, "wadman", "config.json"), nil
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	defer f.Close()

	data := &struct {
		InstallPath string            `json:"install_path"`
		Version     int               `json:"version"`
		Addons      []json.RawMessage `json:"addons"`
	}{}
	err = json.NewDecoder(f).Decode(data)
	if err != nil {
		return nil, err
	}

	if data.Version > configVersion {
		return nil, fmt.Errorf("config file version %d requires a new version of wadman", data.Version)
	}

	if data.Version < 2 {
		// Config version 2 uses the base directory for the install path, instead of the addons directory
		data.InstallPath = filepath.Dir(filepath.Dir(data.InstallPath))
	}

	var addons []*wadman.CurseForgeAddon
	for i := range data.Addons {
		base := wadman.BaseAddon{}
		if err := json.Unmarshal(data.Addons[i], &base); err != nil {
			return nil, err
		}

		inst, err := base.Type.NewInstance()
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data.Addons[i], &inst); err != nil {
			return nil, err
		}

		addons = append(addons, inst)
	}

	return &Config{
		InstallPath: data.InstallPath,
		Addons:      addons,
	}, nil
}

func SaveConfig(path string, config *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0755)); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	data := &struct {
		InstallPath string                    `json:"install_path"`
		Version     int                       `json:"version"`
		Addons      []*wadman.CurseForgeAddon `json:"addons"`
	}{
		config.InstallPath,
		configVersion,
		config.Addons,
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}
