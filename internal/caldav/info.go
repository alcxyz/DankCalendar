package caldav

import "encoding/xml"

const propfindInfo = `<?xml version="1.0"?>
<d:propfind xmlns:d="DAV:">
  <d:prop>
    <d:displayname/>
    <d:current-user-privilege-set/>
  </d:prop>
</d:propfind>`

// CalendarInfo retrieves the display name and read-only status of a calendar.
func (c *Client) CalendarInfo(calURL string) (name string, readOnly bool) {
	readOnly = true

	data, _, err := c.Propfind(calURL, propfindInfo, "0")
	if err != nil {
		return "", true
	}

	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return "", true
	}

	for _, r := range ms.Responses {
		for _, ps := range r.Propstat {
			if ps.Prop.DisplayName != "" {
				name = ps.Prop.DisplayName
			}
		}
	}

	// Check for write privileges in the raw XML
	// The struct-based approach misses deeply nested privilege elements,
	// so re-parse with a targeted search.
	readOnly = !hasWritePrivilege(data)

	return name, readOnly
}

func hasWritePrivilege(data []byte) bool {
	type privilege struct {
		XMLName xml.Name
	}
	type privilegeSet struct {
		Privileges []privilege `xml:",any"`
	}
	type privWrap struct {
		Set privilegeSet `xml:"current-user-privilege-set"`
	}
	type privPropstat struct {
		Prop privWrap `xml:"prop"`
	}
	type privResponse struct {
		Propstat []privPropstat `xml:"propstat"`
	}
	type privMultistatus struct {
		Responses []privResponse `xml:"response"`
	}

	// This is complex nested XML. Just search for known write privilege local names.
	xmlStr := string(data)
	for _, p := range []string{"write", "write-content", "bind"} {
		if containsElement(xmlStr, p) {
			return true
		}
	}
	return false
}

func containsElement(xml, localName string) bool {
	// Check for <d:write/>, <write/>, <D:write/>, etc.
	patterns := []string{
		"<" + localName + "/>",
		"<" + localName + ">",
		":" + localName + "/>",
		":" + localName + ">",
	}
	for _, p := range patterns {
		if len(xml) > 0 {
			for i := 0; i <= len(xml)-len(p); i++ {
				if xml[i:i+len(p)] == p {
					return true
				}
			}
		}
	}
	return false
}
