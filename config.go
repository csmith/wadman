package main

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
			return &Config{}, nil
		}
		return nil, err
	}

	defer f.Close()

	config := &Config{}
	err = json.NewDecoder(f).Decode(config)
	return config, err
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
