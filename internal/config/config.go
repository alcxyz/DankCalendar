package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Calendar struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	Name       string `json:"name,omitempty"`
	Auth       string `json:"auth,omitempty"`
	Provider   string `json:"provider,omitempty"`
	Account    string `json:"account,omitempty"`
	CalendarID string `json:"calendarId,omitempty"`
	ReadOnly   bool   `json:"readOnly,omitempty"`
}

type GoogleAccount struct {
	ClientID string `json:"clientId"`
}

type Config struct {
	Timezone       string                   `json:"timezone"`
	Calendars      []Calendar               `json:"calendars"`
	GoogleAccounts map[string]GoogleAccount `json:"googleAccounts,omitempty"`
}

func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "dankcalendar")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "dankcalendar")
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
