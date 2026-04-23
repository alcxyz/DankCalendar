package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alcxyz/dankcal/internal/caldav"
	"github.com/alcxyz/dankcal/internal/config"
	"github.com/alcxyz/dankcal/internal/keyring"
	"github.com/alcxyz/dankcal/internal/output"
)

func cmdSetup(args []string) {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	url := fs.String("url", "", "CalDAV server URL")
	user := fs.String("username", "", "CalDAV username")
	fs.Parse(args)

	if !keyring.Available() {
		exitError("secret-tool is not installed — install libsecret-tools")
	}

	reader := bufio.NewReader(os.Stdin)

	if *url == "" {
		fmt.Fprint(os.Stderr, "CalDAV URL: ")
		line, _ := reader.ReadString('\n')
		*url = strings.TrimSpace(line)
	}
	if *user == "" {
		fmt.Fprint(os.Stderr, "Username: ")
		line, _ := reader.ReadString('\n')
		*user = strings.TrimSpace(line)
	}

	fmt.Fprint(os.Stderr, "Password: ")
	pw, _ := reader.ReadString('\n')
	pw = strings.TrimSpace(pw)

	if *url == "" || *user == "" || pw == "" {
		exitError("URL, username, and password are all required")
	}

	// Store password in keyring
	if err := keyring.Store(*user, pw); err != nil {
		exitError("keyring: " + err.Error())
	}

	// Discover calendars
	client, err := caldav.NewClient(*url, *user, pw)
	if err != nil {
		exitError(err.Error())
	}

	discovered, err := client.Discover()
	if err != nil {
		exitError("discovery: " + err.Error())
	}

	if len(discovered) == 0 {
		exitError("no calendars found at " + *url)
	}

	// Detect timezone
	tz := detectTimezone()

	// Build config
	var cals []config.Calendar
	for _, d := range discovered {
		cals = append(cals, config.Calendar{
			URL:      d.URL,
			Username: *user,
		})
	}

	cfg := &config.Config{
		Timezone:  tz,
		Calendars: cals,
	}

	if err := config.Save(cfg); err != nil {
		exitError("save config: " + err.Error())
	}

	names := make([]string, len(discovered))
	for i, d := range discovered {
		names[i] = d.Name
	}

	output.JSON(map[string]any{
		"success":   true,
		"calendars": names,
	})
}

func detectTimezone() string {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz
	}
	// Try reading /etc/timezone
	data, err := os.ReadFile("/etc/timezone")
	if err == nil {
		if tz := strings.TrimSpace(string(data)); tz != "" {
			return tz
		}
	}
	// Try reading /etc/localtime symlink
	target, err := os.Readlink("/etc/localtime")
	if err == nil {
		if idx := strings.Index(target, "zoneinfo/"); idx >= 0 {
			return target[idx+len("zoneinfo/"):]
		}
	}
	return "UTC"
}
