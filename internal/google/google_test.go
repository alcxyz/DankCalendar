package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCalDAVEventsURL(t *testing.T) {
	got := CalDAVEventsURL("user@example.com")
	want := "https://apidata.googleusercontent.com/caldav/v2/user@example.com/events"
	if got != want {
		t.Fatalf("CalDAVEventsURL = %q, want %q", got, want)
	}

	got = CalDAVEventsURL("team calendar@example.com")
	want = "https://apidata.googleusercontent.com/caldav/v2/team%20calendar@example.com/events"
	if got != want {
		t.Fatalf("CalDAVEventsURL escaped = %q, want %q", got, want)
	}
}

func TestAuthCodeURLIncludesPKCEAndState(t *testing.T) {
	c := NewClient("user@example.com", "client-id")
	raw := c.authCodeURL("http://127.0.0.1:1234/callback", "state-123", "challenge-123")

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	checks := map[string]string{
		"client_id":             "client-id",
		"redirect_uri":          "http://127.0.0.1:1234/callback",
		"response_type":         "code",
		"scope":                 CalendarScope,
		"access_type":           "offline",
		"prompt":                "consent",
		"state":                 "state-123",
		"code_challenge":        "challenge-123",
		"code_challenge_method": "S256",
		"login_hint":            "user@example.com",
	}
	for key, want := range checks {
		if got := q.Get(key); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestDiscoverCalendarsMapsCalendarList(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/users/me/calendarList" {
			t.Fatalf("path = %q, want /users/me/calendarList", r.URL.Path)
		}
		if got := r.URL.Query().Get("maxResults"); got != "250" {
			t.Fatalf("maxResults = %q, want 250", got)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "primary@example.com", "summary": "Primary", "accessRole": "owner"},
				{"id": "ro@example.com", "summary": "Read only", "accessRole": "reader"},
				{"id": "deleted@example.com", "summary": "Deleted", "accessRole": "owner", "deleted": true},
			},
		})
	}))
	defer srv.Close()

	c := NewClient("user@example.com", "client-id")
	c.APIEndpoint = srv.URL
	c.HTTPClient = srv.Client()

	cals, err := c.discoverCalendars(context.Background(), "token-123")
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer token-123" {
		t.Fatalf("Authorization = %q, want bearer token", gotAuth)
	}
	if len(cals) != 2 {
		t.Fatalf("len(cals) = %d, want 2", len(cals))
	}
	if cals[0].ID != "primary@example.com" || cals[0].ReadOnly {
		t.Fatalf("first calendar = %+v, want writable primary", cals[0])
	}
	if cals[1].ID != "ro@example.com" || !cals[1].ReadOnly {
		t.Fatalf("second calendar = %+v, want read-only", cals[1])
	}
}
