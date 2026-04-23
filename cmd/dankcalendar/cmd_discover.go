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

	// Build and save config
	var cals []config.Calendar
	for _, d := range discovered {
		cals = append(cals, config.Calendar{
			URL:      d.URL,
			Username: *user,
		})
	}

	cfg := &config.Config{
		Timezone:  tz,
		Calendars: cals,
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
