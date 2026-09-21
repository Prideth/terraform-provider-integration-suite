package http

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestUserAgent(t *testing.T) {
	got := UserAgent("1.2.3")
	want := "terraform-provider-sap-integration-suite/1.2.3"
	if got != want {
		t.Errorf("UserAgent() = %q, want %q", got, want)
	}
}

func TestReadLimited(t *testing.T) {
	body := io.NopCloser(strings.NewReader("hello"))
	data, err := ReadLimited(body)
	if err != nil {
		t.Fatalf("ReadLimited() error: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("ReadLimited() = %q, want %q", data, "hello")
	}
}

func TestReadLimited_RejectsOversizedBody(t *testing.T) {
	body := io.NopCloser(&infiniteReader{})
	if _, err := ReadLimited(body); err == nil {
		t.Fatal("expected an error for a body exceeding MaxResponseBytes")
	}
}

type infiniteReader struct{}

func (r *infiniteReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

func TestClient_Do_SetsUserAgent(t *testing.T) {
	var gotUserAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{UserAgent: "test-agent/1.0"})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	resp.Body.Close()

	if gotUserAgent != "test-agent/1.0" {
		t.Errorf("User-Agent = %q, want test-agent/1.0", gotUserAgent)
	}
}

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

// flakyTransport fails the first failCount requests with a network-level
// error before succeeding, simulating a dropped connection.
type flakyTransport struct {
	failCount int
	attempts  int
}

func (t *flakyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.attempts++
	if t.attempts <= t.failCount {
		return nil, &net.OpError{Op: "dial", Err: errConnRefused{}}
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
}

type errConnRefused struct{}

func (errConnRefused) Error() string { return "connection refused" }

func TestClient_RetriesNetworkLevelErrors(t *testing.T) {
	transport := &flakyTransport{failCount: 2}
	client := New(Config{
		Transport:  &http.Client{Transport: transport},
		BaseDelay:  time.Millisecond,
		MaxDelay:   5 * time.Millisecond,
		MaxRetries: 4,
	})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.invalid", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	resp.Body.Close()

	if transport.attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", transport.attempts)
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
