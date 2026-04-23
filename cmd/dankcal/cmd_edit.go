package main

import (
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/alcxyz/dankcal/internal/caldav"
	"github.com/alcxyz/dankcal/internal/config"
	"github.com/alcxyz/dankcal/internal/ical"
	"github.com/alcxyz/dankcal/internal/keyring"
	"github.com/alcxyz/dankcal/internal/output"
)

func cmdEdit(args []string) {
	fs := flag.NewFlagSet("edit", flag.ExitOnError)
	calIdx := fs.Int("calendar", 0, "calendar index")
	filename := fs.String("filename", "", "event ICS filename (e.g. uuid.ics)")
	title := fs.String("title", "", "new event title")
	startDate := fs.String("start-date", "", "new start date (YYYYMMDD)")
	startTime := fs.String("start-time", "", "new start time (HHMM)")
	endDate := fs.String("end-date", "", "new end date (YYYYMMDD)")
	endTime := fs.String("end-time", "", "new end time (HHMM)")
	location := fs.String("location", "", "new location (use empty string to clear)")
	allDay := fs.Bool("all-day", false, "convert to all-day event")
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

	// Build event URL
	calURL := cal.URL
	if calURL[len(calURL)-1] != '/' {
		calURL += "/"
	}
	base, _ := url.Parse(calURL)
	ref, _ := url.Parse(*filename)
	eventURL := base.ResolveReference(ref).String()

	// GET current ICS
	icsBytes, err := client.Get(eventURL)
	if err != nil {
		exitError("fetch event: " + err.Error())
	}
	icsData := string(icsBytes)

	// Apply modifications
	if *title != "" {
		icsData = replaceICSLine(icsData, "SUMMARY", ical.EscapeICS(*title))
	}
	if *startDate != "" && *startTime != "" {
		newDT := fmt.Sprintf("DTSTART;TZID=%s:%sT%s00", cfg.Timezone, *startDate, *startTime)
		icsData = replaceICSPrefixed(icsData, "DTSTART", newDT)
	} else if *startDate != "" && *allDay {
		newDT := fmt.Sprintf("DTSTART;VALUE=DATE:%s", *startDate)
		icsData = replaceICSPrefixed(icsData, "DTSTART", newDT)
	}
	if *endDate != "" && *endTime != "" {
		newDT := fmt.Sprintf("DTEND;TZID=%s:%sT%s00", cfg.Timezone, *endDate, *endTime)
		icsData = replaceICSPrefixed(icsData, "DTEND", newDT)
	} else if *endDate != "" && *allDay {
		newDT := fmt.Sprintf("DTEND;VALUE=DATE:%s", *endDate)
		icsData = replaceICSPrefixed(icsData, "DTEND", newDT)
	}
	if fs.Lookup("location").Value.String() != "" || *location != "" {
		if hasICSLine(icsData, "LOCATION") {
			icsData = replaceICSLine(icsData, "LOCATION", ical.EscapeICS(*location))
		} else if *location != "" {
			icsData = strings.Replace(icsData, "END:VEVENT",
				fmt.Sprintf("LOCATION:%s\nEND:VEVENT", ical.EscapeICS(*location)), 1)
		}
	}

	// PUT back
	status, err := client.Put(eventURL, []byte(icsData))
	if err != nil {
		exitError("PUT: " + err.Error())
	}

	if status >= 200 && status < 300 {
		output.JSON(map[string]any{"success": true})
	} else {
		msg := fmt.Sprintf("server returned %d", status)
		if status == 403 {
			msg = "this calendar is read-only"
		}
		output.JSON(map[string]any{"success": false, "error": msg})
	}
}

// replaceICSLine replaces the value of a simple ICS property line.
func replaceICSLine(ics, propName, newValue string) string {
	var result []string
	for _, line := range strings.Split(ics, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, propName+":") {
			result = append(result, propName+":"+newValue)
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// replaceICSPrefixed replaces a line starting with propName (which may have params).
func replaceICSPrefixed(ics, propName, newLine string) string {
	var result []string
	for _, line := range strings.Split(ics, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, propName+";") || strings.HasPrefix(trimmed, propName+":") {
			result = append(result, newLine)
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func hasICSLine(ics, propName string) bool {
	for _, line := range strings.Split(ics, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, propName+":") || strings.HasPrefix(trimmed, propName+";") {
			return true
		}
	}
	return false
}
