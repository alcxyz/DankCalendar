package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/ical"
	"github.com/alcxyz/DankCalendar/internal/output"
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
	description := fs.String("description", "", "new description (use empty string to clear)")
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
	if cal.ReadOnly {
		output.JSON(map[string]any{"success": false, "error": "this calendar is read-only"})
		return
	}

	client, err := newCalendarClient(context.Background(), cfg, cal)
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
		icsData = setICSProperty(icsData, "SUMMARY", ical.EscapeICS(*title))
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
	if flagProvided(fs, "location") {
		icsData = setICSProperty(icsData, "LOCATION", ical.EscapeICS(*location))
	}
	if flagProvided(fs, "description") {
		icsData = setICSProperty(icsData, "DESCRIPTION", ical.EscapeICS(*description))
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

func setICSProperty(ics, propName, escapedValue string) string {
	if hasICSLine(ics, propName) {
		return replaceICSPrefixed(ics, propName, propName+":"+escapedValue)
	}
	if escapedValue == "" {
		return ics
	}
	return strings.Replace(ics, "END:VEVENT", propName+":"+escapedValue+"\nEND:VEVENT", 1)
}

func flagProvided(fs *flag.FlagSet, name string) bool {
	provided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			provided = true
		}
	})
	return provided
}

// replaceICSPrefixed replaces a line starting with propName (which may have params).
func replaceICSPrefixed(ics, propName, newLine string) string {
	var result []string
	skipFolded := false
	for _, line := range strings.Split(ics, "\n") {
		if skipFolded {
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				continue
			}
			skipFolded = false
		}

		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, propName+";") || strings.HasPrefix(trimmed, propName+":") {
			result = append(result, newLine)
			skipFolded = true
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
