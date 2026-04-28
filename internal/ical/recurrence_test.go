package ical

import (
	"testing"
	"time"
)

func TestAdjustRecurrence_Yearly(t *testing.T) {
	// Birthday on March 15, created in 2020
	ev := &Event{
		Start:  "2020-03-15",
		End:    "2020-03-16",
		AllDay: true,
		RRule:  "FREQ=YEARLY",
	}
	rangeStart := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC)

	if !AdjustRecurrence(ev, rangeStart, rangeEnd) {
		t.Fatal("expected occurrence in range")
	}
	if ev.Start != "2026-03-15" {
		t.Errorf("Start = %q, want %q", ev.Start, "2026-03-15")
	}
	if ev.End != "2026-03-16" {
		t.Errorf("End = %q, want %q", ev.End, "2026-03-16")
	}
}

func TestAdjustRecurrence_YearlyNotInRange(t *testing.T) {
	// Birthday on December 25, but we're looking at April
	ev := &Event{
		Start:  "2020-12-25",
		End:    "2020-12-26",
		AllDay: true,
		RRule:  "FREQ=YEARLY",
	}
	rangeStart := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	if AdjustRecurrence(ev, rangeStart, rangeEnd) {
		t.Errorf("expected no occurrence in range, got Start=%q", ev.Start)
	}
}

func TestAdjustRecurrence_Monthly(t *testing.T) {
	// Monthly reminder on the 30th, created Jan 2024
	ev := &Event{
		Start:  "2024-01-30",
		End:    "2024-01-31",
		AllDay: true,
		RRule:  "FREQ=MONTHLY",
	}
	rangeStart := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	if !AdjustRecurrence(ev, rangeStart, rangeEnd) {
		t.Fatal("expected occurrence in range")
	}
	if ev.Start != "2026-04-30" {
		t.Errorf("Start = %q, want %q", ev.Start, "2026-04-30")
	}
}

func TestAdjustRecurrence_Weekly(t *testing.T) {
	// Weekly event on Fridays, created 2024
	ev := &Event{
		Start: "2024-04-26T20:00:00",
		End:   "2024-04-26T21:00:00",
		RRule: "FREQ=WEEKLY",
	}
	rangeStart := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	if !AdjustRecurrence(ev, rangeStart, rangeEnd) {
		t.Fatal("expected occurrence in range")
	}
	if ev.Start != "2026-05-01T20:00:00" {
		t.Errorf("Start = %q, want %q", ev.Start, "2026-05-01T20:00:00")
	}
	if ev.End != "2026-05-01T21:00:00" {
		t.Errorf("End = %q, want %q", ev.End, "2026-05-01T21:00:00")
	}
}

func TestAdjustRecurrence_WithUntil(t *testing.T) {
	// Yearly event that expired in 2025
	ev := &Event{
		Start:  "2020-06-01",
		End:    "2020-06-02",
		AllDay: true,
		RRule:  "FREQ=YEARLY;UNTIL=20251231",
	}
	rangeStart := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)

	if AdjustRecurrence(ev, rangeStart, rangeEnd) {
		t.Error("expected no occurrence — RRULE expired")
	}
}

func TestAdjustRecurrence_NoRRule(t *testing.T) {
	ev := &Event{
		Start:  "2022-04-30",
		AllDay: true,
	}
	rangeStart := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	if AdjustRecurrence(ev, rangeStart, rangeEnd) {
		t.Error("expected false for event without RRULE")
	}
}

func TestAdjustRecurrence_WithInterval(t *testing.T) {
	// Biweekly event
	ev := &Event{
		Start: "2024-04-26T20:00:00",
		End:   "2024-04-26T21:00:00",
		RRule: "FREQ=WEEKLY;INTERVAL=2",
	}
	rangeStart := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)

	if !AdjustRecurrence(ev, rangeStart, rangeEnd) {
		t.Fatal("expected occurrence in range")
	}
	// The occurrence should respect the 2-week interval from the original date
	t.Logf("Adjusted start: %s", ev.Start)
}
