package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alcxyz/DankCalendar/internal/caldav"
	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/ical"
	"github.com/alcxyz/DankCalendar/internal/keyring"
	"github.com/alcxyz/DankCalendar/internal/output"
)

func cmdNotify(args []string) {
	fs := flag.NewFlagSet("notify", flag.ExitOnError)
	minutes := fs.Int("before", 15, "minutes before event to notify")
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		exitError("config: " + err.Error())
	}

	now := time.Now()
	end := now.Add(time.Duration(*minutes) * time.Minute)

	var allEvents []ical.Event
	for i, cal := range cfg.Calendars {
		pw, err := keyring.Lookup(cal.Username)
		if err != nil {
			continue
		}
		client, err := caldav.NewClient(cal.URL, cal.Username, pw)
		if err != nil {
			continue
		}
		results, err := client.ListEvents(cal.URL, now, end)
		if err != nil {
			continue
		}
		for _, r := range results {
			ev := ical.ParseVEvent(r.ICSData, r.Href, i)
			if ev != nil && !ev.AllDay {
				allEvents = append(allEvents, *ev)
			}
		}
	}

	notified := loadNotified()
	count := 0

	for _, ev := range allEvents {
		key := ev.Summary + "|" + ev.Start
		if _, seen := notified[key]; seen {
			continue
		}

		// Build notification body
		var body []string
		if t := extractTime(ev.Start); t != "" {
			endT := extractTime(ev.End)
			if endT != "" && endT != t {
				body = append(body, t+" - "+endT)
			} else {
				body = append(body, t)
			}
		}
		if ev.Location != "" {
			body = append(body, ev.Location)
		}

		exec.Command("notify-send", "-i", "calendar", "-u", "normal",
			"-a", "DankCalendar", ev.Summary, strings.Join(body, "\n")).Run()

		notified[key] = now.Format(time.RFC3339)
		count++
	}

	if count > 0 {
		saveNotified(notified)
	}

	output.JSON(map[string]any{"notified": count})
}

func extractTime(isoStr string) string {
	if idx := strings.IndexByte(isoStr, 'T'); idx >= 0 {
		t := isoStr[idx+1:]
		if len(t) >= 5 {
			return t[:5]
		}
	}
	return ""
}

func notifyStatePath() string {
	cache := os.Getenv("XDG_CACHE_HOME")
	if cache == "" {
		home, _ := os.UserHomeDir()
		cache = filepath.Join(home, ".cache")
	}
	return filepath.Join(cache, "dankcalendar", "notified.json")
}

func loadNotified() map[string]string {
	data, err := os.ReadFile(notifyStatePath())
	if err != nil {
		return make(map[string]string)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return make(map[string]string)
	}
	// Prune entries older than 24h
	cutoff := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	for k, v := range m {
		if v < cutoff {
			delete(m, k)
		}
	}
	return m
}

func saveNotified(m map[string]string) {
	p := notifyStatePath()
	os.MkdirAll(filepath.Dir(p), 0700)
	data, _ := json.Marshal(m)
	os.WriteFile(p, data, 0600)
}

