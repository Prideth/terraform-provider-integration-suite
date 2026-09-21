package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_RetriesTransientStatusCodes(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected final status 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestClient_DoesNotRetryClientErrors(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("expected exactly 1 attempt for a 404, got %d", got)
	}
}

func TestClient_HonorsRetryAfterHeader(t *testing.T) {
	var attempts int32
	start := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer resp.Body.Close()

	if elapsed := time.Since(start); elapsed < time.Second {
		t.Fatalf("expected the client to wait at least the Retry-After duration, waited %s", elapsed)
	}
}

func TestClient_StopsRetryingWhenContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: 50 * time.Millisecond, MaxDelay: 200 * time.Millisecond, MaxRetries: 10})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	_, err = client.Do(req)
	if err == nil {
		t.Fatal("expected an error once the context was cancelled")
	}
}

func TestParseRetryAfter(t *testing.T) {
	cases := map[string]time.Duration{
		"":     0,
		"5":    5 * time.Second,
		"-1":   0,
		"abcd": 0,
	}
	for input, want := range cases {
		if got := parseRetryAfter(input); got != want {
			t.Errorf("parseRetryAfter(%q) = %s, want %s", input, got, want)
		}
	}
}
