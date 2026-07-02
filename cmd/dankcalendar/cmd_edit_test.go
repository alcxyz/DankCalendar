package main

import (
	"strings"
	"testing"
)

func TestSetICSPropertyInsertsDescription(t *testing.T) {
	ics := "BEGIN:VCALENDAR\nBEGIN:VEVENT\nUID:one\nEND:VEVENT\nEND:VCALENDAR\n"

	got := setICSProperty(ics, "DESCRIPTION", "Bring notes")

	if !strings.Contains(got, "DESCRIPTION:Bring notes\nEND:VEVENT") {
		t.Fatalf("description was not inserted before END:VEVENT:\n%s", got)
	}
}

func TestSetICSPropertyReplacesFoldedDescription(t *testing.T) {
	ics := "BEGIN:VCALENDAR\nBEGIN:VEVENT\nUID:one\nDESCRIPTION:Old detail\n continued detail\nEND:VEVENT\nEND:VCALENDAR\n"

	got := setICSProperty(ics, "DESCRIPTION", "New detail")

	if !strings.Contains(got, "DESCRIPTION:New detail\nEND:VEVENT") {
		t.Fatalf("description was not replaced:\n%s", got)
	}
	if strings.Contains(got, "continued detail") {
		t.Fatalf("folded continuation was not removed:\n%s", got)
	}
}

func TestSetICSPropertyClearsExistingDescription(t *testing.T) {
	ics := "BEGIN:VCALENDAR\nBEGIN:VEVENT\nUID:one\nDESCRIPTION:Old detail\nEND:VEVENT\nEND:VCALENDAR\n"

	got := setICSProperty(ics, "DESCRIPTION", "")

	if !strings.Contains(got, "DESCRIPTION:\nEND:VEVENT") {
		t.Fatalf("description was not cleared:\n%s", got)
	}
}
