package main

import (
	"flag"
	"sort"
	"time"

	"github.com/alcxyz/DankCalendar/internal/caldav"
	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/ical"
	"github.com/alcxyz/DankCalendar/internal/keyring"
	"github.com/alcxyz/DankCalendar/internal/output"
)

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	days := fs.Int("days", 7, "number of days to look ahead")
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		exitError("config: " + err.Error())
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, *days)

	var allEvents []ical.Event

	for i, cal := range cfg.Calendars {
		pw, err := keyring.Lookup(cal.Username)
		if err != nil {
			exitError(err.Error())
		}

		client, err := caldav.NewClient(cal.URL, cal.Username, pw)
		if err != nil {
			exitError(err.Error())
		}

		results, err := client.ListEvents(cal.URL, start, end)
		if err != nil {
			exitError("list events: " + err.Error())
		}

		for _, r := range results {
			ev := ical.ParseVEvent(r.ICSData, r.Href, i)
			if ev != nil {
				allEvents = append(allEvents, *ev)
			}
		}
	}

	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Start < allEvents[j].Start
	})

	output.JSON(map[string]any{
		"events": allEvents,
		"count":  len(allEvents),
	})
}
