package google

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alcxyz/DankCalendar/internal/keyring"
)

const (
	AuthURL        = "https://accounts.google.com/o/oauth2/v2/auth"
	TokenURL       = "https://oauth2.googleapis.com/token"
	CalendarAPI    = "https://www.googleapis.com/calendar/v3"
	CalendarScope  = "https://www.googleapis.com/auth/calendar"
	KeyringService = "dankcalendar-google-oauth"
)

type Calendar struct {
	ID         string
	Name       string
	AccessRole string
	ReadOnly   bool
}

type Client struct {
	Account       string
	ClientID      string
	HTTPClient    *http.Client
	OpenBrowser   func(string) error
	TokenEndpoint string
	APIEndpoint   string
}

func NewClient(account, clientID string) *Client {
	return &Client{
		Account:       account,
		ClientID:      clientID,
		HTTPClient:    &http.Client{Timeout: 15 * time.Second},
		OpenBrowser:   openBrowser,
		TokenEndpoint: TokenURL,
		APIEndpoint:   CalendarAPI,
	}
}

func (c *Client) Authorize(ctx context.Context) error {
	verifier, err := randomURLString(32)
	if err != nil {
		return err
	}
	state, err := randomURLString(32)
	if err != nil {
		return err
	}
	challenge := codeChallenge(verifier)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen for OAuth callback: %w", err)
	}
	defer listener.Close()

	redirectURI := "http://" + listener.Addr().String() + "/callback"
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if got := q.Get("state"); got != state {
				http.Error(w, "invalid state", http.StatusBadRequest)
				errCh <- fmt.Errorf("OAuth callback state mismatch")
				return
			}
			if oauthErr := q.Get("error"); oauthErr != "" {
				http.Error(w, oauthErr, http.StatusBadRequest)
				errCh <- fmt.Errorf("OAuth error: %s", oauthErr)
				return
			}
			code := q.Get("code")
			if code == "" {
				http.Error(w, "missing code", http.StatusBadRequest)
				errCh <- fmt.Errorf("OAuth callback missing code")
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, "<html><body><h1>DankCalendar authorized</h1><p>You can close this tab.</p></body></html>")
			codeCh <- code
		}),
	}
	defer server.Shutdown(context.Background())

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := c.authCodeURL(redirectURI, state, challenge)
	fmt.Fprintf(os.Stderr, "Open this URL to authorize DankCalendar:\n%s\n", authURL)
	if c.OpenBrowser != nil {
		_ = c.OpenBrowser(authURL)
	}

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}

	token, err := c.exchangeCode(ctx, code, redirectURI, verifier)
	if err != nil {
		return err
	}
	if token.RefreshToken == "" {
		return fmt.Errorf("Google did not return a refresh token; revoke access and authorize again")
	}
	return StoreRefreshToken(c.Account, token.RefreshToken)
}

func (c *Client) AccessToken(ctx context.Context) (string, error) {
	refreshToken, err := LookupRefreshToken(c.Account)
	if err != nil {
		return "", err
	}

	form := url.Values{
		"client_id":     {c.ClientID},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}
	token, err := c.postToken(ctx, form)
	if err != nil {
		return "", err
	}
	if token.AccessToken == "" {
		return "", fmt.Errorf("Google token response did not include access_token")
	}
	return token.AccessToken, nil
}

func (c *Client) DiscoverCalendars(ctx context.Context) ([]Calendar, error) {
	accessToken, err := c.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return c.discoverCalendars(ctx, accessToken)
}

func (c *Client) discoverCalendars(ctx context.Context, accessToken string) ([]Calendar, error) {
	var calendars []Calendar
	pageToken := ""
	for {
		u, err := url.Parse(strings.TrimRight(c.APIEndpoint, "/") + "/users/me/calendarList")
		if err != nil {
			return nil, err
		}
		q := u.Query()
		q.Set("maxResults", "250")
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("Google calendarList: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var parsed calendarListResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("parse Google calendarList: %w", err)
		}
		for _, item := range parsed.Items {
			if item.Deleted {
				continue
			}
			name := item.Summary
			if name == "" {
				name = item.ID
			}
			cal := Calendar{
				ID:         item.ID,
				Name:       name,
				AccessRole: item.AccessRole,
				ReadOnly:   item.AccessRole == "reader" || item.AccessRole == "freeBusyReader",
			}
			calendars = append(calendars, cal)
		}
		if parsed.NextPageToken == "" {
			break
		}
		pageToken = parsed.NextPageToken
	}
	return calendars, nil
}

func StoreRefreshToken(account, refreshToken string) error {
	return keyring.StoreService(
		KeyringService,
		account,
		fmt.Sprintf("dankcalendar Google OAuth (%s)", account),
		refreshToken,
	)
}

func LookupRefreshToken(account string) (string, error) {
	return keyring.LookupService(KeyringService, account)
}

func CalDAVEventsURL(calendarID string) string {
	return "https://apidata.googleusercontent.com/caldav/v2/" + url.PathEscape(calendarID) + "/events"
}

func (c *Client) authCodeURL(redirectURI, state, challenge string) string {
	q := url.Values{
		"client_id":             {c.ClientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {CalendarScope},
		"access_type":           {"offline"},
		"prompt":                {"consent"},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	if c.Account != "" {
		q.Set("login_hint", c.Account)
	}
	return AuthURL + "?" + q.Encode()
}

func (c *Client) exchangeCode(ctx context.Context, code, redirectURI, verifier string) (*tokenResponse, error) {
	form := url.Values{
		"client_id":     {c.ClientID},
		"code":          {code},
		"code_verifier": {verifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}
	return c.postToken(ctx, form)
}

func (c *Client) postToken(ctx context.Context, form url.Values) (*tokenResponse, error) {
	endpoint := c.TokenEndpoint
	if endpoint == "" {
		endpoint = TokenURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Google token endpoint: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var token tokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("parse Google token response: %w", err)
	}
	return &token, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type calendarListResponse struct {
	NextPageToken string              `json:"nextPageToken"`
	Items         []calendarListEntry `json:"items"`
}

type calendarListEntry struct {
	ID         string `json:"id"`
	Summary    string `json:"summary"`
	AccessRole string `json:"accessRole"`
	Hidden     bool   `json:"hidden"`
	Deleted    bool   `json:"deleted"`
}

func randomURLString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func openBrowser(rawURL string) error {
	for _, name := range []string{"xdg-open", "gio"} {
		if _, err := exec.LookPath(name); err != nil {
			continue
		}
		if name == "gio" {
			return exec.Command(name, "open", rawURL).Start()
		}
		return exec.Command(name, rawURL).Start()
	}
	return nil
}
