package ical

import (
	"strings"
	"testing"
	"time"
)

func TestEscapeICS(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"a, b, c", "a\\, b\\, c"},
		{"semi;colon", "semi\\;colon"},
		{"back\\slash", "back\\\\slash"},
		{"line\none", "line\\none"},
	}
	for _, tt := range tests {
		got := EscapeICS(tt.in)
		if got != tt.want {
			t.Errorf("EscapeICS(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBuildVEvent_Timed(t *testing.T) {
	start := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)
	ics := BuildVEvent("uid-1", "Meeting", "Room 1", "Europe/Lisbon", start, end, false)

	if !strings.Contains(ics, "UID:uid-1") {
		t.Error("missing UID")
	}
	if !strings.Contains(ics, "SUMMARY:Meeting") {
		t.Error("missing SUMMARY")
	}
	if !strings.Contains(ics, "LOCATION:Room 1") {
		t.Error("missing LOCATION")
	}
	if !strings.Contains(ics, "DTSTART;TZID=Europe/Lisbon:20260424T100000") {
		t.Error("wrong DTSTART")
	}
	if !strings.Contains(ics, "DTEND;TZID=Europe/Lisbon:20260424T110000") {
		t.Error("wrong DTEND")
	}
	if !strings.Contains(ics, "PRODID:-//DankCalendar") {
		t.Error("missing PRODID")
	}
}

func TestBuildVEvent_AllDay(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	ics := BuildVEvent("uid-2", "Holiday", "", "UTC", start, end, true)

	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20260501") {
		t.Error("wrong DTSTART for all-day")
	}
	if !strings.Contains(ics, "DTEND;VALUE=DATE:20260502") {
		t.Error("wrong DTEND for all-day")
	}
	if strings.Contains(ics, "LOCATION:") {
		t.Error("should not have LOCATION when empty")
	}
}

func TestBuildVEvent_EscapedSummary(t *testing.T) {
	start := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)
	ics := BuildVEvent("uid-3", "Test, with; special", "", "UTC", start, end, false)

	if !strings.Contains(ics, "SUMMARY:Test\\, with\\; special") {
		t.Errorf("summary not escaped properly in: %s", ics)
	}
}

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"abc123.ics", false},
		{"a0b1c2d3-e4f5-6789-abcd-ef0123456789.ics", false},
		{"", true},
		{"../evil.ics", true},
		{"foo/bar.ics", true},
		{"test.txt", true},
		{"foo\\bar.ics", true},
		{string([]byte{0x00}) + ".ics", true},
	}
	for _, tt := range tests {
		err := ValidateFilename(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateFilename(%q): err=%v, wantErr=%v", tt.name, err, tt.wantErr)
		}
	}
}
