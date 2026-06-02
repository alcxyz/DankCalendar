package main

import (
	"context"

	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/output"
)

type calendarInfo struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	ReadOnly bool   `json:"readOnly"`
}

func cmdCalendars(args []string) {
	cfg, err := config.Load()
	if err != nil {
		exitError("config: " + err.Error())
	}

	var cals []calendarInfo

	for i, cal := range cfg.Calendars {
		if cal.Auth == "google-oauth" && cal.Name != "" {
			cals = append(cals, calendarInfo{
				Index:    i,
				Name:     cal.Name,
				URL:      cal.URL,
				ReadOnly: cal.ReadOnly,
			})
			continue
		}

		client, err := newCalendarClient(context.Background(), cfg, cal)
		if err != nil {
			exitError(err.Error())
		}

		name, readOnly := client.CalendarInfo(cal.URL)
		cals = append(cals, calendarInfo{
			Index:    i,
			Name:     name,
			URL:      cal.URL,
			ReadOnly: readOnly,
		})
	}

	output.JSON(map[string]any{"calendars": cals})
}
