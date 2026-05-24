package auth

import (
	"net/http"
	"testing"
)

func TestBasicAuthorize(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := (Basic{Username: "user", Password: "pass"}).Authorize(req); err != nil {
		t.Fatal(err)
	}

	const want = "Basic dXNlcjpwYXNz"
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestBearerAuthorize(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := (Bearer{Token: "abc123"}).Authorize(req); err != nil {
		t.Fatal(err)
	}

	const want = "Bearer abc123"
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}
