package main

import (
	"context"
	"fmt"

	"github.com/alcxyz/DankCalendar/internal/auth"
	"github.com/alcxyz/DankCalendar/internal/caldav"
	"github.com/alcxyz/DankCalendar/internal/config"
	"github.com/alcxyz/DankCalendar/internal/google"
	"github.com/alcxyz/DankCalendar/internal/keyring"
)

var lookupBasicPassword = keyring.Lookup

var lookupGoogleAccessToken = func(ctx context.Context, account, clientID string) (string, error) {
	return google.NewClient(account, clientID).AccessToken(ctx)
}

func newCalendarClient(ctx context.Context, cfg *config.Config, cal config.Calendar) (*caldav.Client, error) {
	switch cal.Auth {
	case "", "basic":
		pw, err := lookupBasicPassword(cal.Username)
		if err != nil {
			return nil, err
		}
		return caldav.NewClient(cal.URL, cal.Username, pw)
	case "google-oauth":
		account := cal.Account
		if account == "" {
			account = cal.Username
		}
		gacct, ok := cfg.GoogleAccounts[account]
		if !ok || gacct.ClientID == "" {
			return nil, fmt.Errorf("missing Google OAuth client ID for account %q", account)
		}
		token, err := lookupGoogleAccessToken(ctx, account, gacct.ClientID)
		if err != nil {
			return nil, err
		}
		return caldav.NewClientWithAuth(cal.URL, cal.Username, "", auth.Bearer{Token: token})
	default:
		return nil, fmt.Errorf("unsupported auth type %q for %s", cal.Auth, cal.Username)
	}
}
