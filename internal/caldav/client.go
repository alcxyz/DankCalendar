package caldav

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	BaseURL  *url.URL
	Username string
	Password string
	http     *http.Client
}

func NewClient(rawURL, username, password string) (*Client, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("HTTPS required, got %q", u.Scheme)
	}
	return &Client{
		BaseURL:  u,
		Username: username,
		Password: password,
		http:     &http.Client{},
	}, nil
}

func (c *Client) do(method, rawURL, body string, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}
	cred := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
	req.Header.Set("Authorization", "Basic "+cred)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.http.Do(req)
}

func (c *Client) Propfind(rawURL, body, depth string) ([]byte, string, error) {
	headers := map[string]string{
		"Content-Type": "application/xml; charset=utf-8",
		"Depth":        depth,
	}

	currentURL := rawURL
	for i := 0; i < 5; i++ {
		resp, err := c.do("PROPFIND", currentURL, body, headers)
		if err != nil {
			return nil, currentURL, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 301 || resp.StatusCode == 302 ||
			resp.StatusCode == 307 || resp.StatusCode == 308 {
			loc := resp.Header.Get("Location")
			if loc != "" {
				currentURL = loc
				continue
			}
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, currentURL, err
		}

		if resp.StatusCode >= 400 {
			return nil, currentURL, fmt.Errorf("PROPFIND %s: %d", currentURL, resp.StatusCode)
		}

		return data, currentURL, nil
	}
	return nil, currentURL, fmt.Errorf("too many redirects from %s", rawURL)
}

func (c *Client) Report(calURL, body string) ([]byte, error) {
	headers := map[string]string{
		"Content-Type": "application/xml; charset=utf-8",
		"Depth":        "1",
	}
	resp, err := c.do("REPORT", calURL, body, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("REPORT %s: %d", calURL, resp.StatusCode)
	}
	return data, nil
}

func (c *Client) Put(rawURL string, icsData []byte) (int, error) {
	headers := map[string]string{
		"Content-Type": "text/calendar; charset=utf-8",
	}
	resp, err := c.do("PUT", rawURL, string(icsData), headers)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *Client) Delete(rawURL string) (int, error) {
	resp, err := c.do("DELETE", rawURL, "", nil)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *Client) Get(rawURL string) ([]byte, error) {
	resp, err := c.do("GET", rawURL, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GET %s: %d", rawURL, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// ResolveURL safely resolves a relative path against the base URL.
func (c *Client) ResolveURL(path string) string {
	ref, err := url.Parse(path)
	if err != nil {
		return path
	}
	return c.BaseURL.ResolveReference(ref).String()
}
