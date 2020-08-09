package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Addon struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	FileId      int      `json:"file_id"`
	Directories []string `json:"directories"`
}

type Config struct {
	InstallPath string   `json:"install_path"`
	Version     int      `json:"version"`
	Addons      []*Addon `json:"addons"`
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
			return &Config{
				Version: 2,
			}, nil
		}
		return nil, err
	}

	defer f.Close()

	config := &Config{}
	err = json.NewDecoder(f).Decode(config)
	if err != nil {
		return nil, err
	}

	if config.Version < 2 {
		// Config version 2 uses the base directory for the install path, instead of the addons directory
		config.InstallPath = filepath.Dir(filepath.Dir(config.InstallPath))
		config.Version = 2
	}

	return config, nil
}

func SaveConfig(path string, config *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0755)); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(config); err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}
