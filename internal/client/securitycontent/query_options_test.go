package securitycontent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A tenant probe (September 2026) showed that these entity sets reject query
// options: OAuth2ClientCredentials and UserCredentials answer $top and $select
// with 501, and KeystoreEntries answers every option, even $format=json, with
// 400. Only the bare request succeeds. Every read in this package must
// therefore send no query string at all; this test keeps it that way.
func TestReads_SendNoQueryOptions(t *testing.T) {
	var queries []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			queries = append(queries, r.URL.Path+"?"+r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/KeystoreEntries" {
			_, _ = w.Write([]byte(`{"d":{"results":[]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"d":{}}`))
	}))
	defer server.Close()

	client := New(http.DefaultClient, server.URL)
	ctx := context.Background()

	reads := map[string]func() error{
		"ListKeystoreEntries": func() error { _, err := client.ListKeystoreEntries(ctx); return err },
		"GetKeystoreEntry":    func() error { _, err := client.GetKeystoreEntry(ctx, "alias"); return err },
		"GetUserCredential":   func() error { _, err := client.GetUserCredential(ctx, "cred"); return err },
		"GetOAuth2ClientCredential": func() error {
			_, err := client.GetOAuth2ClientCredential(ctx, "cred")
			return err
		},
	}
	for name, read := range reads {
		if err := read(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	for _, q := range queries {
		t.Errorf("read sent query options SAP rejects: %s", q)
	}
}
