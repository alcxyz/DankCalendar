package auth

import (
	"encoding/base64"
	"net/http"
)

type Provider interface {
	Authorize(req *http.Request) error
}

type Basic struct {
	Username string
	Password string
}

func (b Basic) Authorize(req *http.Request) error {
	cred := base64.StdEncoding.EncodeToString([]byte(b.Username + ":" + b.Password))
	req.Header.Set("Authorization", "Basic "+cred)
	return nil
}

type Bearer struct {
	Token string
}

func (b Bearer) Authorize(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+b.Token)
	return nil
}
