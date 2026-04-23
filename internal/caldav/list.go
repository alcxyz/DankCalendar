package caldav

import (
	"encoding/xml"
	"fmt"
	"time"
)

// ListEvents fetches events from a calendar URL within the given time range.
// Returns raw ICS strings for each event found.
func (c *Client) ListEvents(calURL string, start, end time.Time) ([]EventResult, error) {
	body := fmt.Sprintf(`<?xml version="1.0"?>
<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:prop>
    <d:getetag/>
    <c:calendar-data/>
  </d:prop>
  <c:filter>
    <c:comp-filter name="VCALENDAR">
      <c:comp-filter name="VEVENT">
        <c:time-range start="%s" end="%s"/>
      </c:comp-filter>
    </c:comp-filter>
  </c:filter>
</c:calendar-query>`, start.UTC().Format("20060102T150405Z"), end.UTC().Format("20060102T150405Z"))

	data, err := c.Report(calURL, body)
	if err != nil {
		return nil, err
	}

	return extractEvents(data, calURL)
}

type EventResult struct {
	Href     string
	ETag     string
	ICSData  string
}

type reportMultistatus struct {
	XMLName   xml.Name         `xml:"multistatus"`
	Responses []reportResponse `xml:"response"`
}

type reportResponse struct {
	Href     string           `xml:"href"`
	Propstat []reportPropstat `xml:"propstat"`
}

type reportPropstat struct {
	Prop   reportProp `xml:"prop"`
	Status string     `xml:"status"`
}

type reportProp struct {
	ETag         string `xml:"getetag"`
	CalendarData string `xml:"calendar-data"`
}

func extractEvents(data []byte, calURL string) ([]EventResult, error) {
	var ms reportMultistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil, fmt.Errorf("parse REPORT response: %w", err)
	}

	var results []EventResult
	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.CalendarData == "" {
				continue
			}
			results = append(results, EventResult{
				Href:    r.Href,
				ETag:    ps.Prop.ETag,
				ICSData: ps.Prop.CalendarData,
			})
		}
	}
	return results, nil
}
