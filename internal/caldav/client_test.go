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
