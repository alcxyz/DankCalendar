package caldav

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

type DiscoveredCalendar struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Discover finds all calendars on a CalDAV server by following the
// well-known -> principal -> calendar-home-set -> calendar listing chain.
func (c *Client) Discover() ([]DiscoveredCalendar, error) {
	base := strings.TrimRight(c.BaseURL.String(), "/")

	// Step 1: Try .well-known/caldav at the host root (RFC 5785/6764).
	// .well-known URIs must be at the server root, not appended to the path.
	effectiveBase := c.BaseURL.Scheme + "://" + c.BaseURL.Host
	wellKnown := effectiveBase + "/.well-known/caldav"
	_, redirected, err := c.Propfind(wellKnown, propfindPrincipal, "0")
	if err == nil {
		if u, err := url.Parse(redirected); err == nil {
			effectiveBase = u.Scheme + "://" + u.Host
		}
	}

	// Step 2: Find current-user-principal at the provided URL.
	data, _, err := c.Propfind(base+"/", propfindPrincipal, "0")
	if err != nil {
		return nil, fmt.Errorf("principal lookup: %w", err)
	}

	// If current-user-principal is not reported (e.g. Google Calendar returns
	// <unauthenticated/> or omits it), treat the provided URL as the principal.
	principalURL := base + "/"
	if principalHref, err := extractPrincipalHref(data); err == nil {
		principalURL = resolveHref(effectiveBase, principalHref)
	}

	// Step 3: Find calendar-home-set
	data, _, err = c.Propfind(principalURL, propfindHomeSet, "0")
	if err != nil {
		return nil, fmt.Errorf("home-set lookup: %w", err)
	}

	homeSetHref, err := extractHomeSetHref(data)
	if err != nil {
		return nil, err
	}

	homeSetURL := resolveHref(effectiveBase, homeSetHref)

	// Step 4: List calendars in home-set
	data, _, err = c.Propfind(homeSetURL, propfindCalendars, "1")
	if err != nil {
		return nil, fmt.Errorf("calendar listing: %w", err)
	}

	return extractCalendars(data, effectiveBase)
}

func resolveHref(base, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	return base + href
}

// XML parsing for multistatus responses

const propfindPrincipal = `<?xml version="1.0"?>
<d:propfind xmlns:d="DAV:">
  <d:prop>
    <d:current-user-principal/>
  </d:prop>
</d:propfind>`

const propfindHomeSet = `<?xml version="1.0"?>
<d:propfind xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:prop>
    <c:calendar-home-set/>
  </d:prop>
</d:propfind>`

const propfindCalendars = `<?xml version="1.0"?>
<d:propfind xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav"
  xmlns:ic="http://apple.com/ns/ical/">
  <d:prop>
    <d:displayname/>
    <d:resourcetype/>
    <ic:calendar-color/>
  </d:prop>
</d:propfind>`

// We use a lenient XML approach: unmarshal into generic structures
// and walk them, since CalDAV servers vary in namespace handling.

type multistatus struct {
	XMLName   xml.Name   `xml:"multistatus"`
	Responses []response `xml:"response"`
}

type response struct {
	Href     string    `xml:"href"`
	Propstat []propstat `xml:"propstat"`
}

type propstat struct {
	Prop   prop   `xml:"prop"`
	Status string `xml:"status"`
}

type prop struct {
	DisplayName  string       `xml:"displayname"`
	ResourceType resourceType `xml:"resourcetype"`
	Principal    principal    `xml:"current-user-principal"`
	HomeSet      homeSet      `xml:"calendar-home-set"`
}

type resourceType struct {
	Calendar *struct{} `xml:"calendar"`
}

type principal struct {
	Href string `xml:"href"`
}

type homeSet struct {
	Href string `xml:"href"`
}

func extractPrincipalHref(data []byte) (string, error) {
	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return "", fmt.Errorf("parse principal response: %w", err)
	}
	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.Principal.Href != "" {
				return ps.Prop.Principal.Href, nil
			}
		}
	}
	return "", fmt.Errorf("could not find current-user-principal")
}

func extractHomeSetHref(data []byte) (string, error) {
	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return "", fmt.Errorf("parse home-set response: %w", err)
	}
	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.HomeSet.Href != "" {
				return ps.Prop.HomeSet.Href, nil
			}
		}
	}
	return "", fmt.Errorf("could not find calendar-home-set")
}

func extractCalendars(data []byte, base string) ([]DiscoveredCalendar, error) {
	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil, fmt.Errorf("parse calendar list: %w", err)
	}

	var cals []DiscoveredCalendar
	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.ResourceType.Calendar == nil {
				continue
			}
			name := ps.Prop.DisplayName
			if name == "" {
				parts := strings.Split(strings.TrimRight(r.Href, "/"), "/")
				name = parts[len(parts)-1]
			}
			fullURL := resolveHref(base, r.Href)
			cals = append(cals, DiscoveredCalendar{URL: fullURL, Name: name})
		}
	}
	return cals, nil
}
