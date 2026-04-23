package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/alcxyz/DankCalendar/internal/caldav"
	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/ical"
	"github.com/alcxyz/DankCalendar/internal/keyring"
	"github.com/alcxyz/DankCalendar/internal/output"
)

func cmdDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	calIdx := fs.Int("calendar", 0, "calendar index")
	filename := fs.String("filename", "", "event ICS filename (e.g. uuid.ics)")
	fs.Parse(args)

	if *filename == "" {
		exitError("--filename is required")
	}
	if err := ical.ValidateFilename(*filename); err != nil {
		exitError(err.Error())
	}

	cfg, err := config.Load()
	if err != nil {
		exitError("config: " + err.Error())
	}
	if *calIdx >= len(cfg.Calendars) {
		exitError(fmt.Sprintf("invalid calendar index %d", *calIdx))
	}

	cal := cfg.Calendars[*calIdx]
	pw, err := keyring.Lookup(cal.Username)
	if err != nil {
		exitError(err.Error())
	}

	client, err := caldav.NewClient(cal.URL, cal.Username, pw)
	if err != nil {
		exitError(err.Error())
	}

	calURL := cal.URL
	if calURL[len(calURL)-1] != '/' {
		calURL += "/"
	}
	base, _ := url.Parse(calURL)
	ref, _ := url.Parse(*filename)
	eventURL := base.ResolveReference(ref).String()

	status, err := client.Delete(eventURL)
	if err != nil {
		exitError("DELETE: " + err.Error())
	}

	if status >= 200 && status < 300 {
		output.JSON(map[string]any{"success": true})
	} else {
		msg := fmt.Sprintf("server returned %d", status)
		switch status {
		case 403:
			msg = "this calendar is read-only"
		case 404:
			msg = "event not found"
		case 401:
			msg = "authentication failed"
		}
		output.JSON(map[string]any{"success": false, "error": msg})
	}
}
