package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Calendar struct {
	URL      string `json:"url"`
	Username string `json:"username"`
}

type Config struct {
	Timezone string     `json:"timezone"`
	Calendars []Calendar `json:"calendars"`
}

func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "dankcal")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "dankcal")
}

func Path() string {
	return filepath.Join(Dir(), "config.json")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0600)
}
