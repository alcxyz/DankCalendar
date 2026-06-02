package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/google"
	"github.com/alcxyz/DankCalendar/internal/keyring"
	"github.com/alcxyz/DankCalendar/internal/output"
)

func cmdGoogleDiscover(args []string) {
	fs := flag.NewFlagSet("google-discover", flag.ExitOnError)
	account := fs.String("account", "", "Google account email")
	clientID := fs.String("client-id", "", "Google OAuth desktop client ID")
	forceAuthorize := fs.Bool("authorize", false, "run the browser OAuth flow before discovery")
	fs.Parse(args)

	if *account == "" {
		exitError("--account is required")
	}
	if !keyring.Available() {
		exitError("secret-tool is not installed — install libsecret-tools")
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{Timezone: detectTimezone()}
	}
	if cfg.GoogleAccounts == nil {
		cfg.GoogleAccounts = make(map[string]config.GoogleAccount)
	}
	if *clientID == "" {
		*clientID = cfg.GoogleAccounts[*account].ClientID
	}
	if *clientID == "" {
		exitError("--client-id is required for first Google setup")
	}

	gclient := google.NewClient(*account, *clientID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if _, err := google.LookupRefreshToken(*account); err != nil || *forceAuthorize {
		if err := gclient.Authorize(ctx); err != nil {
			exitError("Google authorization: " + err.Error())
		}
	}

	discovered, err := gclient.DiscoverCalendars(ctx)
	if err != nil {
		exitError("Google discovery: " + err.Error())
	}
	if len(discovered) == 0 {
		output.JSON(map[string]any{
			"success":   false,
			"error":     "no calendars found",
			"calendars": []string{},
		})
		return
	}

	names := applyGoogleDiscovery(cfg, *account, *clientID, discovered)

	if err := config.Save(cfg); err != nil {
		exitError("save config: " + err.Error())
	}

	output.JSON(map[string]any{
		"success":   true,
		"calendars": names,
	})
}

func applyGoogleDiscovery(cfg *config.Config, account, clientID string, discovered []google.Calendar) []string {
	if cfg.GoogleAccounts == nil {
		cfg.GoogleAccounts = make(map[string]config.GoogleAccount)
	}
	cfg.GoogleAccounts[account] = config.GoogleAccount{ClientID: clientID}

	var kept []config.Calendar
	for _, cal := range cfg.Calendars {
		if cal.Auth == "google-oauth" && cal.Account == account {
			continue
		}
		kept = append(kept, cal)
	}
	for _, cal := range discovered {
		kept = append(kept, config.Calendar{
			URL:        google.CalDAVEventsURL(cal.ID),
			Username:   account,
			Name:       cal.Name,
			Auth:       "google-oauth",
			Provider:   "google",
			Account:    account,
			CalendarID: cal.ID,
			ReadOnly:   cal.ReadOnly,
		})
	}
	cfg.Calendars = kept

	names := make([]string, len(discovered))
	for i, cal := range discovered {
		names[i] = cal.Name
		if cal.ReadOnly {
			names[i] = fmt.Sprintf("%s (read-only)", cal.Name)
		}
	}
	return names
}
