package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var version = "dev"

var commands = map[string]string{
	"list":      "List upcoming events",
	"calendars": "Discover available calendars",
	"add":       "Create a new event",
	"edit":      "Modify an existing event",
	"delete":    "Delete an event",
	"notify":    "Send desktop notifications for upcoming events",
	"setup":     "Configure CalDAV credentials",
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "--help", "-h":
		usage()
	case "--version", "-v":
		fmt.Println(version)
	case "list":
		exitError("not yet implemented: list")
	case "calendars":
		exitError("not yet implemented: calendars")
	case "add":
		exitError("not yet implemented: add")
	case "edit":
		exitError("not yet implemented: edit")
	case "delete":
		exitError("not yet implemented: delete")
	case "notify":
		exitError("not yet implemented: notify")
	case "setup":
		exitError("not yet implemented: setup")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: dankcal <command> [flags]\n\nCommands:\n")
	for name, desc := range commands {
		fmt.Fprintf(os.Stderr, "  %-12s %s\n", name, desc)
	}
	fmt.Fprintf(os.Stderr, "\nFlags:\n  --help       Show this help\n  --version    Print version\n")
}

func exitError(msg string) {
	out, _ := json.Marshal(map[string]string{"error": msg})
	fmt.Println(string(out))
	os.Exit(1)
}
