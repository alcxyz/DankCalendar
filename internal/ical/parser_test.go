package ical

import "testing"

func TestParseVEvent_Timed(t *testing.T) {
	ics := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test-123
SUMMARY:Team Meeting
DTSTART;TZID=Europe/Lisbon:20260424T100000
DTEND;TZID=Europe/Lisbon:20260424T110000
LOCATION:Room 42
DESCRIPTION:Weekly sync
END:VEVENT
END:VCALENDAR`

	ev := ParseVEvent(ics, "/cal/test-123.ics", 0, "Europe/Lisbon")
	if ev == nil {
		t.Fatal("expected event, got nil")
	}
	if ev.UID != "test-123" {
		t.Errorf("UID = %q, want %q", ev.UID, "test-123")
	}
	if ev.Summary != "Team Meeting" {
		t.Errorf("Summary = %q, want %q", ev.Summary, "Team Meeting")
	}
	if ev.Start != "2026-04-24T10:00:00" {
		t.Errorf("Start = %q, want %q", ev.Start, "2026-04-24T10:00:00")
	}
	if ev.End != "2026-04-24T11:00:00" {
		t.Errorf("End = %q, want %q", ev.End, "2026-04-24T11:00:00")
	}
	if ev.AllDay {
		t.Error("AllDay should be false")
	}
	if ev.Location != "Room 42" {
		t.Errorf("Location = %q, want %q", ev.Location, "Room 42")
	}
	if ev.Filename != "test-123.ics" {
		t.Errorf("Filename = %q, want %q", ev.Filename, "test-123.ics")
	}
}

func TestParseVEvent_AllDay(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:allday-1
SUMMARY:Holiday
DTSTART;VALUE=DATE:20260501
DTEND;VALUE=DATE:20260502
END:VEVENT
END:VCALENDAR`

	ev := ParseVEvent(ics, "/cal/allday-1.ics", 1, "Europe/Lisbon")
	if ev == nil {
		t.Fatal("expected event, got nil")
	}
	if ev.Start != "2026-05-01" {
		t.Errorf("Start = %q, want %q", ev.Start, "2026-05-01")
	}
	if ev.End != "2026-05-02" {
		t.Errorf("End = %q, want %q", ev.End, "2026-05-02")
	}
	if !ev.AllDay {
		t.Error("AllDay should be true")
	}
	if ev.CalendarIdx != 1 {
		t.Errorf("CalendarIdx = %d, want 1", ev.CalendarIdx)
	}
}

func TestParseVEvent_FoldedLines(t *testing.T) {
	ics := "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:fold-1\r\nSUMMARY:Very long\r\n  event title here\r\nDTSTART;VALUE=DATE:20260601\r\nEND:VEVENT\r\nEND:VCALENDAR"

	ev := ParseVEvent(ics, "", 0, "")
	if ev == nil {
		t.Fatal("expected event, got nil")
	}
	if ev.Summary != "Very long event title here" {
		t.Errorf("Summary = %q, want %q", ev.Summary, "Very long event title here")
	}
}

func TestParseVEvent_EscapedChars(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:esc-1
SUMMARY:Meeting\, with\; special\\chars
DESCRIPTION:Line one\nLine two\NLine three
DTSTART;VALUE=DATE:20260601
END:VEVENT
END:VCALENDAR`

	ev := ParseVEvent(ics, "", 0, "")
	if ev == nil {
		t.Fatal("expected event, got nil")
	}
	if ev.Summary != "Meeting, with; special\\chars" {
		t.Errorf("Summary = %q", ev.Summary)
	}
	if ev.Description != "Line one\nLine two\nLine three" {
		t.Errorf("Description = %q", ev.Description)
	}
}

func TestParseVEvent_UTCTimestamp(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:utc-1
SUMMARY:UTC Event
DTSTART:20260424T140000Z
DTEND:20260424T150000Z
END:VEVENT
END:VCALENDAR`

	// Target is Lisbon (UTC+1 in summer / WEST), so 14:00Z → 15:00
	ev := ParseVEvent(ics, "", 0, "Europe/Lisbon")
	if ev == nil {
		t.Fatal("expected event, got nil")
	}
	if ev.Start != "2026-04-24T15:00:00" {
		t.Errorf("Start = %q, want %q", ev.Start, "2026-04-24T15:00:00")
	}
	if ev.AllDay {
		t.Error("AllDay should be false for UTC datetime")
	}
}

func TestParseVEvent_NoEndAllDay(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:noend-1
SUMMARY:Single Day
DTSTART;VALUE=DATE:20260701
END:VEVENT
END:VCALENDAR`

	ev := ParseVEvent(ics, "", 0, "")
	if ev == nil {
		t.Fatal("expected event, got nil")
	}
	if ev.End != "2026-07-02" {
		t.Errorf("End = %q, want %q (default next day)", ev.End, "2026-07-02")
	}
}

func TestParseVEvent_UTCToLocalConversion(t *testing.T) {
	// A UTC event and a TZID event that represent the same moment
	// should produce identical local times when parsed to the same target.
	icsUTC := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:cross-1
SUMMARY:UTC Call
DTSTART:20260428T090000Z
END:VEVENT
END:VCALENDAR`

	icsLisbon := `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:cross-2
SUMMARY:Lisbon Call
DTSTART;TZID=Europe/Lisbon:20260428T100000
END:VEVENT
END:VCALENDAR`

	evUTC := ParseVEvent(icsUTC, "", 0, "Europe/Lisbon")
	evLisbon := ParseVEvent(icsLisbon, "", 1, "Europe/Lisbon")

	if evUTC == nil || evLisbon == nil {
		t.Fatal("expected both events, got nil")
	}
	if evUTC.Start != evLisbon.Start {
		t.Errorf("UTC start %q != Lisbon start %q — should be equal after normalisation",
			evUTC.Start, evLisbon.Start)
	}
}
