package ical

import (
	"path"
	"strings"
	"time"
)

type Event struct {
	UID         string `json:"uid"`
	Summary     string `json:"title"`
	Start       string `json:"start"`
	End         string `json:"end"`
	AllDay      bool   `json:"allDay"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Filename    string `json:"filename"`
	CalendarIdx int    `json:"calendarIndex"`
}

// ParseVEvent extracts a single VEVENT from ICS data.
// href is the event's URL path, used to derive the filename.
func ParseVEvent(icsData, href string, calIdx int) *Event {
	lines := unfold(icsData)

	var inEvent bool
	ev := &Event{CalendarIdx: calIdx}

	for _, line := range lines {
		if line == "BEGIN:VEVENT" {
			inEvent = true
			continue
		}
		if line == "END:VEVENT" {
			break
		}
		if !inEvent {
			continue
		}

		name, params, value := parseLine(line)
		switch name {
		case "UID":
			ev.UID = value
		case "SUMMARY":
			ev.Summary = unescapeICS(value)
		case "LOCATION":
			ev.Location = unescapeICS(value)
		case "DESCRIPTION":
			ev.Description = unescapeICS(value)
		case "DTSTART":
			ev.Start, ev.AllDay = parseDateTime(value, params)
		case "DTEND":
			ev.End, _ = parseDateTime(value, params)
		}
	}

	if ev.UID == "" && ev.Summary == "" {
		return nil
	}

	// Derive filename from href
	if href != "" {
		ev.Filename = path.Base(href)
	}

	// Default end for all-day events: start + 1 day
	if ev.AllDay && ev.End == "" {
		if t, err := time.Parse("2006-01-02", ev.Start); err == nil {
			ev.End = t.AddDate(0, 0, 1).Format("2006-01-02")
		}
	}

	return ev
}

// unfold handles ICS line folding (lines starting with space/tab are
// continuations of the previous line).
func unfold(data string) []string {
	var lines []string
	for _, raw := range strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		if len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t') && len(lines) > 0 {
			lines[len(lines)-1] += raw[1:]
		} else {
			lines = append(lines, raw)
		}
	}
	return lines
}

// parseLine splits "NAME;PARAM=VAL:value" into name, params string, value.
func parseLine(line string) (name, params, value string) {
	// Find the first unquoted colon
	colonIdx := -1
	inQuote := false
	for i, ch := range line {
		if ch == '"' {
			inQuote = !inQuote
		}
		if ch == ':' && !inQuote {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return line, "", ""
	}
	value = line[colonIdx+1:]
	nameParams := line[:colonIdx]

	if idx := strings.IndexByte(nameParams, ';'); idx >= 0 {
		return nameParams[:idx], nameParams[idx+1:], value
	}
	return nameParams, "", value
}

// parseDateTime converts ICS date/datetime values to ISO strings.
// Returns the ISO string and whether it's an all-day (DATE-only) value.
func parseDateTime(value, params string) (string, bool) {
	// Check for VALUE=DATE (all-day)
	if strings.Contains(strings.ToUpper(params), "VALUE=DATE") {
		// YYYYMMDD -> YYYY-MM-DD
		if len(value) >= 8 {
			return value[:4] + "-" + value[4:6] + "-" + value[6:8], true
		}
		return value, true
	}

	// Try parsing as datetime: YYYYMMDDTHHMMSS or YYYYMMDDTHHMMSSZ
	value = strings.TrimSuffix(value, "Z")
	if len(value) >= 15 && value[8] == 'T' {
		iso := value[:4] + "-" + value[4:6] + "-" + value[6:8] +
			"T" + value[9:11] + ":" + value[11:13] + ":" + value[13:15]
		return iso, false
	}

	// Fallback: might be just a date
	if len(value) == 8 {
		return value[:4] + "-" + value[4:6] + "-" + value[6:8], true
	}

	return value, false
}

func unescapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\N", "\n")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
