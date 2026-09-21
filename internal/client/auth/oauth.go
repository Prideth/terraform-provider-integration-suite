// Package auth provides OAuth 2.0 client credentials authentication for SAP
// Integration Suite APIs, with token caching and automatic refresh handled by
// golang.org/x/oauth2.
package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Config holds the OAuth 2.0 client credentials needed to authenticate
// against one SAP Integration Suite API area.
type Config struct {
	TokenURL     string
	ClientID     string
	ClientSecret string

	// Scopes is optional; most Integration Suite OAuth clients derive their
	// scopes from the role collections assigned to the service key instead
	// of requesting explicit scopes.
	Scopes []string
}

func (c Config) validate() error {
	if c.TokenURL == "" {
		return fmt.Errorf("oauth: token URL must not be empty")
	}
	if c.ClientID == "" {
		return fmt.Errorf("oauth: client ID must not be empty")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("oauth: client secret must not be empty")
	}
	return nil
}

// TokenSource builds an oauth2.TokenSource that performs the client
// credentials flow against TokenURL, caching the resulting token and
// refreshing it automatically once it is close to expiry. The returned
// source is safe for concurrent use.
//
// base is the underlying *http.Client used to reach the token endpoint (so
// that proxy settings, timeouts, and TLS configuration stay consistent with
// the rest of the provider); it must not be nil.
func (c Config) HTTPClient(ctx context.Context, base *http.Client) (*http.Client, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	cfg := clientcredentials.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		TokenURL:     c.TokenURL,
		Scopes:       c.Scopes,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, base)
	return cfg.Client(ctx), nil
}
