package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOldBasicAuthConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "dankcalendar")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{
  "timezone": "UTC",
  "calendars": [
    {"url": "https://caldav.example.com/user/", "username": "user@example.com"}
  ]
}`)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Timezone != "UTC" {
		t.Fatalf("Timezone = %q, want UTC", cfg.Timezone)
	}
	if len(cfg.Calendars) != 1 {
		t.Fatalf("len(Calendars) = %d, want 1", len(cfg.Calendars))
	}
	cal := cfg.Calendars[0]
	if cal.Auth != "" || cal.Provider != "" || cal.Account != "" {
		t.Fatalf("old config should leave provider metadata empty, got %+v", cal)
	}
}
