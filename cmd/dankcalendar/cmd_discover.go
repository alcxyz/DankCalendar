package main

import (
	"flag"

	"github.com/alcxyz/DankCalendar/internal/caldav"
	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/keyring"
	"github.com/alcxyz/DankCalendar/internal/output"
)

func cmdDiscover(args []string) {
	fs := flag.NewFlagSet("discover", flag.ExitOnError)
	url := fs.String("url", "", "CalDAV server URL")
	user := fs.String("username", "", "CalDAV username")
	pw := fs.String("password", "", "CalDAV password")
	appendMode := fs.Bool("append", false, "append discovered calendars to existing config instead of replacing")
	fs.Parse(args)

	if *url == "" || *user == "" || *pw == "" {
		exitError("--url, --username, and --password are all required")
	}

	// Store password in keyring
	if !keyring.Available() {
		exitError("secret-tool is not installed — install libsecret-tools")
	}
	if err := keyring.Store(*user, *pw); err != nil {
		exitError("keyring: " + err.Error())
	}

	// Discover calendars
	client, err := caldav.NewClient(*url, *user, *pw)
	if err != nil {
		exitError(err.Error())
	}

	discovered, err := client.Discover()
	if err != nil {
		output.JSON(map[string]any{
			"success":   false,
			"error":     err.Error(),
			"calendars": []string{},
		})
		return
	}

	if len(discovered) == 0 {
		output.JSON(map[string]any{
			"success":   false,
			"error":     "no calendars found",
			"calendars": []string{},
		})
		return
	}

	// Detect timezone
	tz := detectTimezone()

	// Build new calendar entries
	var newCals []config.Calendar
	for _, d := range discovered {
		newCals = append(newCals, config.Calendar{
			URL:      d.URL,
			Username: *user,
		})
	}

	// In append mode, merge with existing config (preserving other accounts)
	var cfg *config.Config
	if *appendMode {
		existing, err := config.Load()
		if err == nil {
			// Remove any existing entries for this username, then append fresh ones
			var kept []config.Calendar
			for _, c := range existing.Calendars {
				if c.Username != *user {
					kept = append(kept, c)
				}
			}
			cfg = &config.Config{
				Timezone:  existing.Timezone,
				Calendars: append(kept, newCals...),
			}
		} else {
			cfg = &config.Config{Timezone: tz, Calendars: newCals}
		}
	} else {
		cfg = &config.Config{Timezone: tz, Calendars: newCals}
	}

	if err := config.Save(cfg); err != nil {
		exitError("save config: " + err.Error())
	}

	names := make([]string, len(discovered))
	for i, d := range discovered {
		names[i] = d.Name
	}

	output.JSON(map[string]any{
		"success":   true,
		"calendars": names,
	})
}
