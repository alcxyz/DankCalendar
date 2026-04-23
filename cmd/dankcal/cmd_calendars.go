package main

import (
	"github.com/alcxyz/dankcal/internal/caldav"
	"github.com/alcxyz/dankcal/internal/config"
	"github.com/alcxyz/dankcal/internal/keyring"
	"github.com/alcxyz/dankcal/internal/output"
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
		pw, err := keyring.Lookup(cal.Username)
		if err != nil {
			exitError(err.Error())
		}

		client, err := caldav.NewClient(cal.URL, cal.Username, pw)
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
