package ical

import (
	"strconv"
	"strings"
	"time"
)

// rruleParams parses an RRULE value string into a key/value map.
// e.g. "FREQ=YEARLY;INTERVAL=2;UNTIL=20301231T235959Z" →
//
//	{"FREQ":"YEARLY", "INTERVAL":"2", "UNTIL":"20301231T235959Z"}
func rruleParams(rrule string) map[string]string {
	m := make(map[string]string)
	for _, part := range strings.Split(rrule, ";") {
		k, v, ok := strings.Cut(part, "=")
		if ok {
			m[strings.ToUpper(k)] = v
		}
	}
	return m
}

// AdjustRecurrence shifts an event whose DTSTART is before rangeStart into
// the query window [rangeStart, rangeEnd) using its RRULE.
//
// It updates ev.Start (and ev.End, preserving duration) in place and returns
// true if an occurrence was found within the range.  Returns false if the
// event should be dropped (no occurrence in range, or RRULE expired).
//
// Handles FREQ=YEARLY, MONTHLY, WEEKLY, DAILY with INTERVAL and UNTIL.
func AdjustRecurrence(ev *Event, rangeStart, rangeEnd time.Time) bool {
	if ev.RRule == "" {
		return false
	}

	params := rruleParams(ev.RRule)
	freq := params["FREQ"]
	if freq == "" {
		return false
	}

	interval := 1
	if v, err := strconv.Atoi(params["INTERVAL"]); err == nil && v > 0 {
		interval = v
	}

	// Parse UNTIL if present
	var until time.Time
	if u := params["UNTIL"]; u != "" {
		raw := strings.TrimSuffix(u, "Z")
		if len(raw) >= 8 {
			if t, err := time.Parse("20060102", raw[:8]); err == nil {
				until = t.AddDate(0, 0, 1) // inclusive
			}
		}
	}

	// Parse original start
	origStart := parseEventTime(ev.Start, rangeStart.Location())
	if origStart.IsZero() {
		return false
	}

	// Parse original end to compute duration
	var dur time.Duration
	if ev.End != "" {
		if origEnd := parseEventTime(ev.End, rangeStart.Location()); !origEnd.IsZero() {
			dur = origEnd.Sub(origStart)
		}
	}
	if dur <= 0 && ev.AllDay {
		dur = 24 * time.Hour
	}

	// Find the first occurrence >= rangeStart.
	// Always add from origStart to avoid day-of-month drift with AddDate.
	var candidate time.Time
	switch freq {
	case "YEARLY":
		// Jump close to the target year, then scan
		n := (rangeStart.Year() - origStart.Year()) / interval
		if n < 0 {
			n = 0
		}
		for {
			candidate = origStart.AddDate(n*interval, 0, 0)
			if !candidate.Before(rangeStart) {
				break
			}
			n++
		}
	case "MONTHLY":
		origY, origM, _ := origStart.Date()
		rangeY, rangeM, _ := rangeStart.Date()
		totalOrig := origY*12 + int(origM)
		totalRange := rangeY*12 + int(rangeM)
		n := (totalRange - totalOrig) / interval
		if n < 0 {
			n = 0
		}
		for {
			candidate = origStart.AddDate(0, n*interval, 0)
			if !candidate.Before(rangeStart) {
				break
			}
			n++
		}
	case "WEEKLY":
		days := int(rangeStart.Sub(origStart).Hours()/24) / (7 * interval)
		if days < 0 {
			days = 0
		}
		for {
			candidate = origStart.AddDate(0, 0, days*7*interval)
			if !candidate.Before(rangeStart) {
				break
			}
			days++
		}
	case "DAILY":
		days := int(rangeStart.Sub(origStart).Hours()/24) / interval
		if days < 0 {
			days = 0
		}
		for {
			candidate = origStart.AddDate(0, 0, days*interval)
			if !candidate.Before(rangeStart) {
				break
			}
			days++
		}
	default:
		return false
	}

	// Check bounds
	if !until.IsZero() && candidate.After(until) {
		return false
	}
	if !candidate.Before(rangeEnd) {
		return false
	}

	// Update event
	if ev.AllDay {
		ev.Start = candidate.Format("2006-01-02")
		endDate := candidate.Add(dur)
		ev.End = endDate.Format("2006-01-02")
	} else {
		ev.Start = candidate.Format("2006-01-02T15:04:05")
		endDate := candidate.Add(dur)
		ev.End = endDate.Format("2006-01-02T15:04:05")
	}

	return true
}

// parseEventTime parses an ISO date or datetime string as produced by
// parseDateTime into a time.Time.
func parseEventTime(s string, loc *time.Location) time.Time {
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, loc); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Time{}
}
