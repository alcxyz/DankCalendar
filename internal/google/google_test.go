package google

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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

func TestDiscoverCalendarsHandlesPagination(t *testing.T) {
	var pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("pageToken")
		pages = append(pages, page)
		if page == "" {
			json.NewEncoder(w).Encode(map[string]any{
				"nextPageToken": "next",
				"items": []map[string]any{
					{"id": "one@example.com", "summary": "One", "accessRole": "owner"},
				},
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "two@example.com", "summary": "Two", "accessRole": "writer"},
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
	if strings.Join(pages, ",") != ",next" {
		t.Fatalf("page tokens = %#v, want initial page then next", pages)
	}
	if len(cals) != 2 || cals[0].ID != "one@example.com" || cals[1].ID != "two@example.com" {
		t.Fatalf("calendars = %+v", cals)
	}
}

func TestPostTokenParsesAccessToken(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(body)
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access-123",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer srv.Close()

	c := NewClient("user@example.com", "client-id")
	c.TokenEndpoint = srv.URL
	c.HTTPClient = srv.Client()

	token, err := c.postToken(context.Background(), url.Values{
		"grant_type": {"refresh_token"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access-123" {
		t.Fatalf("AccessToken = %q, want access-123", token.AccessToken)
	}
	if gotBody != "grant_type=refresh_token" {
		t.Fatalf("body = %q, want encoded form", gotBody)
	}
}

func TestPostTokenReturnsHTTPErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad token", http.StatusBadRequest)
	}))
	defer srv.Close()

	c := NewClient("user@example.com", "client-id")
	c.TokenEndpoint = srv.URL
	c.HTTPClient = srv.Client()

	_, err := c.postToken(context.Background(), url.Values{})
	if err == nil || !strings.Contains(err.Error(), "bad token") {
		t.Fatalf("error = %v, want token endpoint error with body", err)
	}
}

func TestAuthorizeRejectsStateMismatch(t *testing.T) {
	c := NewClient("user@example.com", "client-id")
	c.OpenBrowser = func(rawURL string) error {
		u, err := url.Parse(rawURL)
		if err != nil {
			return err
		}
		redirectURI := u.Query().Get("redirect_uri")
		go http.Get(redirectURI + "?state=wrong&code=code-123")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Authorize(ctx)
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("Authorize error = %v, want state mismatch", err)
	}
}
