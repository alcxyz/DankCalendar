package main

import (
	"fmt"
	"os"

	"github.com/alcxyz/dankcal/internal/output"
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
	args := os.Args[2:]

	switch cmd {
	case "--help", "-h":
		usage()
	case "--version", "-v":
		fmt.Println(version)
	case "list":
		cmdList(args)
	case "calendars":
		cmdCalendars(args)
	case "add":
		cmdAdd(args)
	case "edit":
		cmdEdit(args)
	case "delete":
		cmdDelete(args)
	case "notify":
		cmdNotify(args)
	case "setup":
		cmdSetup(args)
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
	output.Error(msg)
	os.Exit(1)
}
