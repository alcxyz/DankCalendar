package ical

import (
	"fmt"
	"strings"
	"time"
)

// EscapeICS escapes a string for use in ICS property values per RFC 5545.
func EscapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// BuildVEvent generates a complete VCALENDAR with a single VEVENT.
func BuildVEvent(uid, summary, location, timezone string, start, end time.Time, allDay bool) string {
	now := time.Now().UTC().Format("20060102T150405Z")
	summary = EscapeICS(summary)

	var dtstart, dtend string
	if allDay {
		dtstart = fmt.Sprintf("DTSTART;VALUE=DATE:%s", start.Format("20060102"))
		dtend = fmt.Sprintf("DTEND;VALUE=DATE:%s", end.Format("20060102"))
	} else {
		dtstart = fmt.Sprintf("DTSTART;TZID=%s:%s", timezone, start.Format("20060102T150405"))
		dtend = fmt.Sprintf("DTEND;TZID=%s:%s", timezone, end.Format("20060102T150405"))
	}

	var locationLine string
	if location != "" {
		locationLine = fmt.Sprintf("LOCATION:%s\n", EscapeICS(location))
	}

	return fmt.Sprintf("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//DankCalendar\nBEGIN:VEVENT\nUID:%s\n%s\n%s\nDTSTAMP:%s\nSUMMARY:%s\n%sEND:VEVENT\nEND:VCALENDAR\n",
		uid, dtstart, dtend, now, summary, locationLine)
}

// ValidateFilename checks that a filename is safe (no path traversal).
func ValidateFilename(name string) error {
	if name == "" {
		return fmt.Errorf("empty filename")
	}
	if strings.Contains(name, "..") || strings.Contains(name, "/") ||
		strings.Contains(name, "\\") || strings.ContainsRune(name, 0) {
		return fmt.Errorf("invalid filename: %q", name)
	}
	if !strings.HasSuffix(name, ".ics") {
		return fmt.Errorf("filename must end with .ics: %q", name)
	}
	return nil
}
