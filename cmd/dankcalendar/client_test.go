package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alcxyz/DankCalendar/internal/config"
)

func TestNewCalendarClientUsesBasicPasswordLookup(t *testing.T) {
	oldLookup := lookupBasicPassword
	t.Cleanup(func() { lookupBasicPassword = oldLookup })

	var gotUsername string
	lookupBasicPassword = func(username string) (string, error) {
		gotUsername = username
		return "pass", nil
	}

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Basic dXNlcjpwYXNz" {
			t.Fatalf("Authorization = %q, want Basic auth", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := newCalendarClient(context.Background(), &config.Config{}, config.Calendar{
		URL:      srv.URL,
		Username: "user",
	})
	if err != nil {
		t.Fatal(err)
	}
	client.SetHTTPClient(srv.Client())
	if _, err := client.Get(srv.URL); err != nil {
		t.Fatal(err)
	}
	if gotUsername != "user" {
		t.Fatalf("lookup username = %q, want user", gotUsername)
	}
}

func TestNewCalendarClientUsesGoogleBearerToken(t *testing.T) {
	oldLookup := lookupGoogleAccessToken
	t.Cleanup(func() { lookupGoogleAccessToken = oldLookup })

	var gotAccount, gotClientID string
	lookupGoogleAccessToken = func(ctx context.Context, account, clientID string) (string, error) {
		gotAccount = account
		gotClientID = clientID
		return "token-123", nil
	}

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("Authorization = %q, want Bearer token", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &config.Config{
		GoogleAccounts: map[string]config.GoogleAccount{
			"user@example.com": {ClientID: "client-id"},
		},
	}
	client, err := newCalendarClient(context.Background(), cfg, config.Calendar{
		URL:      srv.URL,
		Username: "user@example.com",
		Auth:     "google-oauth",
		Account:  "user@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	client.SetHTTPClient(srv.Client())
	if _, err := client.Get(srv.URL); err != nil {
		t.Fatal(err)
	}
	if gotAccount != "user@example.com" || gotClientID != "client-id" {
		t.Fatalf("google lookup = (%q, %q), want account and client ID", gotAccount, gotClientID)
	}
}

func TestNewCalendarClientRejectsMissingGoogleClientID(t *testing.T) {
	_, err := newCalendarClient(context.Background(), &config.Config{}, config.Calendar{
		URL:      "https://example.com/calendar",
		Username: "user@example.com",
		Auth:     "google-oauth",
	})
	if err == nil {
		t.Fatal("expected missing Google client ID error")
	}
}

func TestNewCalendarClientPropagatesTokenError(t *testing.T) {
	oldLookup := lookupGoogleAccessToken
	t.Cleanup(func() { lookupGoogleAccessToken = oldLookup })

	lookupGoogleAccessToken = func(ctx context.Context, account, clientID string) (string, error) {
		return "", errors.New("token failed")
	}

	cfg := &config.Config{
		GoogleAccounts: map[string]config.GoogleAccount{
			"user@example.com": {ClientID: "client-id"},
		},
	}
	_, err := newCalendarClient(context.Background(), cfg, config.Calendar{
		URL:      "https://example.com/calendar",
		Username: "user@example.com",
		Auth:     "google-oauth",
	})
	if err == nil || err.Error() != "token failed" {
		t.Fatalf("error = %v, want token failed", err)
	}
}
