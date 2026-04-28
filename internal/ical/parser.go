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
	RRule       string `json:"-"` // raw RRULE value, not serialised to JSON
}

// ParseVEvent extracts a single VEVENT from ICS data.
// href is the event's URL path, used to derive the filename.
// targetTZ is the IANA timezone name to normalize timed events into
// (e.g. "Europe/Lisbon"). If empty, the system local timezone is used.
func ParseVEvent(icsData, href string, calIdx int, targetTZ string) *Event {
	loc := loadLocation(targetTZ)
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
			ev.Start, ev.AllDay = parseDateTime(value, params, loc)
		case "DTEND":
			ev.End, _ = parseDateTime(value, params, loc)
		case "RRULE":
			ev.RRule = value
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

// loadLocation returns a *time.Location for the given IANA name,
// falling back to the system local timezone on empty or invalid input.
func loadLocation(tz string) *time.Location {
	if tz == "" {
		return time.Now().Location()
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Now().Location()
	}
	return loc
}

// parseDateTime converts ICS date/datetime values to ISO strings normalised
// to targetLoc. Returns the ISO string and whether it's an all-day value.
func parseDateTime(value, params string, targetLoc *time.Location) (string, bool) {
	// Check for VALUE=DATE (all-day) — no timezone conversion needed
	if strings.Contains(strings.ToUpper(params), "VALUE=DATE") {
		if len(value) >= 8 {
			return value[:4] + "-" + value[4:6] + "-" + value[6:8], true
		}
		return value, true
	}

	// Determine the source timezone and strip the Z suffix
	isUTC := strings.HasSuffix(value, "Z")
	raw := strings.TrimSuffix(value, "Z")

	if len(raw) >= 15 && raw[8] == 'T' {
		// Parse the bare datetime digits
		t, err := time.Parse("20060102T150405", raw)
		if err == nil {
			// Assign the correct source timezone
			var srcLoc *time.Location
			if isUTC {
				srcLoc = time.UTC
			} else if tzid := extractTZID(params); tzid != "" {
				if loc, err := time.LoadLocation(tzid); err == nil {
					srcLoc = loc
				}
			}

			if srcLoc != nil {
				// Re-interpret in source tz, then convert to target
				t = time.Date(t.Year(), t.Month(), t.Day(),
					t.Hour(), t.Minute(), t.Second(), 0, srcLoc)
				t = t.In(targetLoc)
			}
			// If no source tz info, the time is already in local/target tz

			return t.Format("2006-01-02T15:04:05"), false
		}
	}

	// Fallback: might be just a date
	if len(raw) == 8 {
		return raw[:4] + "-" + raw[4:6] + "-" + raw[6:8], true
	}

	return value, false
}

// extractTZID pulls the timezone identifier from an ICS parameter string
// like "TZID=Europe/Lisbon".
func extractTZID(params string) string {
	upper := strings.ToUpper(params)
	idx := strings.Index(upper, "TZID=")
	if idx < 0 {
		return ""
	}
	rest := params[idx+5:]
	if end := strings.IndexByte(rest, ';'); end >= 0 {
		return rest[:end]
	}
	return rest
}

func unescapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\N", "\n")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
