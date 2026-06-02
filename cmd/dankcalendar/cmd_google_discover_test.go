package main

import (
	"testing"

	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/google"
)

func TestApplyGoogleDiscoveryReplacesOnlySameAccountCalendars(t *testing.T) {
	cfg := &config.Config{
		Timezone: "UTC",
		GoogleAccounts: map[string]config.GoogleAccount{
			"old@example.com": {ClientID: "old-client"},
		},
		Calendars: []config.Calendar{
			{URL: "https://caldav.example.com/", Username: "basic@example.com"},
			{URL: "https://old-google.example.com/", Username: "user@example.com", Auth: "google-oauth", Account: "user@example.com", CalendarID: "old"},
			{URL: "https://other-google.example.com/", Username: "other@example.com", Auth: "google-oauth", Account: "other@example.com", CalendarID: "other"},
		},
	}

	names := applyGoogleDiscovery(cfg, "user@example.com", "client-id", []google.Calendar{
		{ID: "primary@example.com", Name: "Primary", ReadOnly: false},
		{ID: "readonly@example.com", Name: "Read Only", ReadOnly: true},
	})

	if len(names) != 2 || names[0] != "Primary" || names[1] != "Read Only (read-only)" {
		t.Fatalf("names = %#v", names)
	}
	if got := cfg.GoogleAccounts["user@example.com"].ClientID; got != "client-id" {
		t.Fatalf("client ID = %q, want client-id", got)
	}
	if len(cfg.Calendars) != 4 {
		t.Fatalf("len(Calendars) = %d, want 4", len(cfg.Calendars))
	}
	if cfg.Calendars[0].Username != "basic@example.com" {
		t.Fatalf("first calendar should preserve basic calendar, got %+v", cfg.Calendars[0])
	}
	if cfg.Calendars[1].Account != "other@example.com" {
		t.Fatalf("second calendar should preserve other Google account, got %+v", cfg.Calendars[1])
	}
	newCal := cfg.Calendars[2]
	if newCal.Auth != "google-oauth" || newCal.Provider != "google" || newCal.Account != "user@example.com" {
		t.Fatalf("new calendar metadata = %+v", newCal)
	}
	if newCal.URL != "https://apidata.googleusercontent.com/caldav/v2/primary@example.com/events" {
		t.Fatalf("new calendar URL = %q", newCal.URL)
	}
	if !cfg.Calendars[3].ReadOnly {
		t.Fatalf("read-only calendar metadata not preserved: %+v", cfg.Calendars[3])
	}
}
