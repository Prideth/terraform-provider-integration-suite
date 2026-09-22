package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_WriteWithoutCSRFRequirement_SucceedsOnFirstAttempt(t *testing.T) {
	// The common case today: SAP accepts the write outright, with no
	// X-CSRF-Token exchange at all. This must remain the behavior when the
	// server never asks for a token, exactly as before this feature existed.
	var attempts int32
	var sawToken string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		sawToken = r.Header.Get("X-CSRF-Token")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("expected exactly 1 request when no CSRF token is required, got %d", got)
	}
	if sawToken != "" {
		t.Errorf("X-CSRF-Token = %q, want empty when the client has never fetched one", sawToken)
	}
}

func TestClient_WriteRejectedForCSRF_FetchesAndRetries(t *testing.T) {
	var postAttempts, fetchAttempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			atomic.AddInt32(&fetchAttempts, 1)
			if r.Header.Get("X-CSRF-Token") != "fetch" {
				t.Errorf("CSRF fetch request X-CSRF-Token = %q, want %q", r.Header.Get("X-CSRF-Token"), "fetch")
			}
			w.Header().Set("X-CSRF-Token", "server-issued-token")
			w.Header().Set("Set-Cookie", "JSESSIONID=abc123; Path=/; HttpOnly")
			w.WriteHeader(http.StatusOK)
			return
		}

		n := atomic.AddInt32(&postAttempts, 1)
		if n == 1 {
			// First write attempt: no token cached yet, SAP rejects it.
			if r.Header.Get("X-CSRF-Token") != "" {
				t.Errorf("first write attempt should carry no token, got %q", r.Header.Get("X-CSRF-Token"))
			}
			w.Header().Set("X-CSRF-Token", "Required")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if r.Header.Get("X-CSRF-Token") != "server-issued-token" {
			t.Errorf("retried write X-CSRF-Token = %q, want %q", r.Header.Get("X-CSRF-Token"), "server-issued-token")
		}
		if r.Header.Get("Cookie") != "JSESSIONID=abc123" {
			t.Errorf("retried write Cookie = %q, want %q", r.Header.Get("Cookie"), "JSESSIONID=abc123")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&postAttempts); got != 2 {
		t.Errorf("expected exactly 2 write attempts (original + one CSRF retry), got %d", got)
	}
	if got := atomic.LoadInt32(&fetchAttempts); got != 1 {
		t.Errorf("expected exactly 1 CSRF token fetch, got %d", got)
	}
}

func TestClient_CachesCSRFTokenAcrossRequests(t *testing.T) {
	var fetchAttempts, postAttempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			atomic.AddInt32(&fetchAttempts, 1)
			w.Header().Set("X-CSRF-Token", "cached-token")
			w.WriteHeader(http.StatusOK)
			return
		}

		n := atomic.AddInt32(&postAttempts, 1)
		if n == 1 {
			w.Header().Set("X-CSRF-Token", "Required")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	// First write: rejected once, fetches a token, retries successfully.
	req1, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, nil)
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("first Do() error: %v", err)
	}
	_ = resp1.Body.Close()

	// Second write: the cached token should be sent immediately, with no
	// second fetch and no rejection round-trip.
	req2, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, nil)
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("second Do() error: %v", err)
	}
	_ = resp2.Body.Close()

	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("second write status = %d, want 201", resp2.StatusCode)
	}
	if got := atomic.LoadInt32(&fetchAttempts); got != 1 {
		t.Errorf("expected exactly 1 CSRF fetch across both writes (token reused), got %d", got)
	}
	if got := atomic.LoadInt32(&postAttempts); got != 3 {
		t.Errorf("expected 3 total write attempts (1 rejected + 1 retry + 1 cached-token success), got %d", got)
	}
}

func TestClient_CSRFFetchFailure_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// SAP fetch endpoint itself errors out; no token in the response.
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-CSRF-Token", "Required")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, nil)
	if _, err := client.Do(req); err == nil { //nolint:bodyclose // err != nil below means resp is always nil here
		t.Fatal("expected an error when the CSRF token fetch itself fails")
	}
}

func TestClient_SecondCSRFRejectionAfterRefresh_ReturnedAsIs(t *testing.T) {
	// A freshly issued token is rejected too: retrying again would not help,
	// so the second rejection must be returned to the caller, not retried
	// indefinitely.
	var postAttempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("X-CSRF-Token", "server-issued-token")
			w.WriteHeader(http.StatusOK)
			return
		}
		atomic.AddInt32(&postAttempts, 1)
		w.Header().Set("X-CSRF-Token", "Required")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want the second rejection (403) returned as-is", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&postAttempts); got != 2 {
		t.Errorf("expected exactly 2 write attempts (no further retries), got %d", got)
	}
}

func TestClient_PlainForbidden_IsNotTreatedAsCSRF(t *testing.T) {
	// A 403 without the SAP CSRF signal header is an ordinary authorization
	// failure and must never trigger a token fetch.
	var fetchAttempts, postAttempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			atomic.AddInt32(&fetchAttempts, 1)
		}
		atomic.AddInt32(&postAttempts, 1)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := New(Config{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&fetchAttempts); got != 0 {
		t.Errorf("expected no CSRF fetch for a plain 403, got %d", got)
	}
	if got := atomic.LoadInt32(&postAttempts); got != 1 {
		t.Errorf("expected exactly 1 write attempt for a plain 403, got %d", got)
	}
}

func TestClient_GetNeverCarriesCSRFToken(t *testing.T) {
	client := New(Config{})
	if needsCSRFToken(http.MethodGet) {
		t.Error("GET must never be treated as CSRF-protected")
	}
	if needsCSRFToken(http.MethodHead) {
		t.Error("HEAD must never be treated as CSRF-protected")
	}
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if !needsCSRFToken(m) {
			t.Errorf("%s must be treated as CSRF-protected", m)
		}
	}
	_ = client
}
