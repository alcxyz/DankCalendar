package keyring

import (
	"fmt"
	"os/exec"
	"strings"
)

const service = "dankcalendar"

func Lookup(username string) (string, error) {
	return LookupService(service, username)
}

func LookupService(serviceName, account string) (string, error) {
	cmd := exec.Command("secret-tool", "lookup", "service", serviceName, "account", account)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keyring lookup failed: %w (is secret-tool installed and keyring unlocked?)", err)
	}
	pw := strings.TrimRight(string(out), "\n")
	if pw == "" {
		return "", fmt.Errorf("no secret found in keyring for account %q", account)
	}
	return pw, nil
}

func Store(username, password string) error {
	return StoreService(service, username, fmt.Sprintf("dankcalendar CalDAV (%s)", username), password)
}

func StoreService(serviceName, account, label, secret string) error {
	cmd := exec.Command("secret-tool", "store",
		"--label", label,
		"service", serviceName, "account", account)
	cmd.Stdin = strings.NewReader(secret)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("keyring store failed: %w", err)
	}
	return nil
}

func Available() bool {
	_, err := exec.LookPath("secret-tool")
	return err == nil
}
