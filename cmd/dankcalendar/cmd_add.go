package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/alcxyz/DankCalendar/internal/caldav"
	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/ical"
	"github.com/alcxyz/DankCalendar/internal/keyring"
	"github.com/alcxyz/DankCalendar/internal/output"
)

func cmdAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	calIdx := fs.Int("calendar", 0, "calendar index")
	summary := fs.String("summary", "", "event title")
	location := fs.String("location", "", "event location")
	startStr := fs.String("start", "", "start time (YYYY-MM-DDTHH:MM or YYYY-MM-DD for all-day)")
	endStr := fs.String("end", "", "end time (YYYY-MM-DDTHH:MM or YYYY-MM-DD for all-day)")
	allDay := fs.Bool("all-day", false, "create all-day event")
	fs.Parse(args)

	if *summary == "" || *startStr == "" {
		exitError("--summary and --start are required")
	}

	cfg, err := config.Load()
	if err != nil {
		exitError("config: " + err.Error())
	}
	if *calIdx >= len(cfg.Calendars) {
		exitError(fmt.Sprintf("invalid calendar index %d (have %d)", *calIdx, len(cfg.Calendars)))
	}

	cal := cfg.Calendars[*calIdx]
	pw, err := keyring.Lookup(cal.Username)
	if err != nil {
		exitError(err.Error())
	}

	var start, end time.Time

	if *allDay {
		start, err = time.Parse("2006-01-02", *startStr)
		if err != nil {
			exitError("invalid start date: " + err.Error())
		}
		if *endStr != "" {
			end, err = time.Parse("2006-01-02", *endStr)
			if err != nil {
				exitError("invalid end date: " + err.Error())
			}
		} else {
			end = start.AddDate(0, 0, 1)
		}
	} else {
		start, err = time.ParseInLocation("2006-01-02T15:04", *startStr, time.Now().Location())
		if err != nil {
			exitError("invalid start time: " + err.Error())
		}
		if *endStr != "" {
			end, err = time.ParseInLocation("2006-01-02T15:04", *endStr, time.Now().Location())
			if err != nil {
				exitError("invalid end time: " + err.Error())
			}
		} else {
			end = start.Add(time.Hour)
		}
	}

	uid := generateUID()
	filename := uid + ".ics"
	icsData := ical.BuildVEvent(uid, *summary, *location, cfg.Timezone, start, end, *allDay)

	// Resolve event URL
	calURL := cal.URL
	if calURL[len(calURL)-1] != '/' {
		calURL += "/"
	}
	base, err := url.Parse(calURL)
	if err != nil {
		exitError("invalid calendar URL: " + err.Error())
	}
	ref, _ := url.Parse(filename)
	eventURL := base.ResolveReference(ref).String()

	client, err := caldav.NewClient(cal.URL, cal.Username, pw)
	if err != nil {
		exitError(err.Error())
	}

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
		} else if status == 401 {
			msg = "authentication failed"
		}
		output.JSON(map[string]any{"success": false, "error": msg})
	}
}

func generateUID() string {
	now := time.Now()
	return fmt.Sprintf("%d-%d", now.UnixNano(), os.Getpid())
}
