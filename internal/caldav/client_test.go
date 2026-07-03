package caldav

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alcxyz/DankCalendar/internal/auth"
)

func TestNewClientWithAuthAuthorizesRequests(t *testing.T) {
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client, err := NewClientWithAuth(srv.URL, "user@example.com", "", auth.Bearer{Token: "token-123"})
	if err != nil {
		t.Fatal(err)
	}
	client.http = srv.Client()

	status, err := client.Delete(srv.URL + "/event.ics")
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", status, http.StatusNoContent)
	}
	if gotAuth != "Bearer token-123" {
		t.Fatalf("Authorization = %q, want bearer token", gotAuth)
	}
}

func TestNewClientPreservesBasicAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "user", "pass")
	if err != nil {
		t.Fatal(err)
	}
	client.http = srv.Client()

	if _, err := client.Get(srv.URL + "/event.ics"); err != nil {
		t.Fatal(err)
	}
	const want = "Basic dXNlcjpwYXNz"
	if gotAuth != want {
		t.Fatalf("Authorization = %q, want %q", gotAuth, want)
	}
}

func TestSafeRedirect(t *testing.T) {
	cases := []struct {
		name     string
		current  string
		location string
		want     string
		wantErr  bool
	}{
		{"same host absolute", "https://dav.example.com/cal/", "https://dav.example.com/principals/", "https://dav.example.com/principals/", false},
		{"relative path", "https://dav.example.com/cal/", "/principals/user/", "https://dav.example.com/principals/user/", false},
		{"cross host rejected", "https://dav.example.com/cal/", "https://attacker.example.net/steal", "", true},
		{"downgrade to http rejected", "https://dav.example.com/cal/", "http://dav.example.com/cal/", "", true},
		{"host case-insensitive", "https://DAV.example.com/cal/", "https://dav.example.com/x", "https://dav.example.com/x", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := safeRedirect(tc.current, tc.location)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("safeRedirect(%q, %q) = %q, want %q", tc.current, tc.location, got, tc.want)
			}
		})
	}
}
