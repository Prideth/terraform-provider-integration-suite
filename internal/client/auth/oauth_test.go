package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestConfig_HTTPClient_AcquiresAndCachesToken(t *testing.T) {
	var tokenRequests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenRequests, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "test-token", "token_type": "bearer", "expires_in": 3600}`))
	}))
	defer server.Close()

	cfg := Config{
		TokenURL:     server.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}

	client, err := cfg.HTTPClient(context.Background(), http.DefaultClient)
	if err != nil {
		t.Fatalf("HTTPClient() error: %v", err)
	}

	var apiCalls int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&apiCalls, 1)
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer test-token")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer api.Close()

	for i := 0; i < 3; i++ {
		resp, err := client.Get(api.URL)
		if err != nil {
			t.Fatalf("Get() error: %v", err)
		}
		_ = resp.Body.Close()
	}

	if got := atomic.LoadInt32(&tokenRequests); got != 1 {
		t.Errorf("expected the token to be fetched once and cached, got %d token requests", got)
	}
	if got := atomic.LoadInt32(&apiCalls); got != 3 {
		t.Errorf("expected 3 API calls, got %d", got)
	}
}

func TestConfig_HTTPClient_RejectsIncompleteConfig(t *testing.T) {
	cases := []Config{
		{ClientID: "id", ClientSecret: "secret"},
		{TokenURL: "https://example.com/token", ClientSecret: "secret"},
		{TokenURL: "https://example.com/token", ClientID: "id"},
	}

	for _, cfg := range cases {
		if _, err := cfg.HTTPClient(context.Background(), http.DefaultClient); err == nil {
			t.Errorf("expected an error for incomplete config %+v", cfg)
		}
	}
}
