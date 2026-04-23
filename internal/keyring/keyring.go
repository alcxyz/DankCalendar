package keyring

import (
	"fmt"
	"os/exec"
	"strings"
)

const service = "dankcalendar"

func Lookup(username string) (string, error) {
	cmd := exec.Command("secret-tool", "lookup", "service", service, "account", username)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keyring lookup failed: %w (is secret-tool installed and keyring unlocked?)", err)
	}
	pw := strings.TrimRight(string(out), "\n")
	if pw == "" {
		return "", fmt.Errorf("no password found in keyring for account %q", username)
	}
	return pw, nil
}

func Store(username, password string) error {
	cmd := exec.Command("secret-tool", "store",
		"--label", fmt.Sprintf("dankcalendar CalDAV (%s)", username),
		"service", service, "account", username)
	cmd.Stdin = strings.NewReader(password)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("keyring store failed: %w", err)
	}
	return nil
}

func Available() bool {
	_, err := exec.LookPath("secret-tool")
	return err == nil
}
