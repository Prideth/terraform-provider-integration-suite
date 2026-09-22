// Package http provides the shared HTTP transport used by every SAP
// Integration Suite API client in this provider: retry with backoff and
// jitter, Retry-After handling, a bounded response size, and a stable
// User-Agent. No API-specific (OData, REST) logic lives here.
package http

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

// UserAgent identifies every request this provider makes. version is set by
// the provider at build time via internal/provider.Version.
func UserAgent(version string) string {
	return fmt.Sprintf("terraform-provider-sap-integration-suite/%s", version)
}

// MaxResponseBytes bounds how much of a response body this client will read,
// to protect against unexpectedly large or malicious responses.
const MaxResponseBytes = 64 * 1024 * 1024 // 64 MiB

const (
	defaultMaxRetries = 4
	defaultBaseDelay  = 250 * time.Millisecond
	defaultMaxDelay   = 30 * time.Second
)

// Config configures the retrying HTTP client.
type Config struct {
	// Transport is the underlying http.Client to wrap (already configured
	// with authentication, timeouts, and TLS settings). If nil,
	// http.DefaultClient is used.
	Transport *http.Client

	UserAgent string

	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration

	// InvalidateToken, if set, is called at most once per Do call the first
	// time a response comes back 401, immediately followed by one retry
	// (with no backoff wait, since the fix is a fresh token, not time). A
	// second 401 after that retry is returned to the caller as-is: SAP
	// rejected a freshly issued token, so retrying further would not help.
	InvalidateToken func()
}

// Client is an HTTP client that retries transient SAP API failures.
type Client struct {
	transport       *http.Client
	userAgent       string
	maxRetries      int
	baseDelay       time.Duration
	maxDelay        time.Duration
	invalidateToken func()
	csrf            *csrfCache
}

// New builds a retrying HTTP client from cfg.
func New(cfg Config) *Client {
	transport := cfg.Transport
	if transport == nil {
		transport = http.DefaultClient
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}
	baseDelay := cfg.BaseDelay
	if baseDelay <= 0 {
		baseDelay = defaultBaseDelay
	}
	maxDelay := cfg.MaxDelay
	if maxDelay <= 0 {
		maxDelay = defaultMaxDelay
	}

	return &Client{
		transport:       transport,
		userAgent:       cfg.UserAgent,
		maxRetries:      maxRetries,
		baseDelay:       baseDelay,
		maxDelay:        maxDelay,
		invalidateToken: cfg.InvalidateToken,
		csrf:            &csrfCache{},
	}
}

// isRetryable reports whether an HTTP status code represents a transient
// failure worth retrying. 401 is handled separately by callers that can
// refresh a token; it is never retried blindly here, and 400/403/404 are
// never retried.
func isRetryable(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// Do executes req, retrying transient failures with exponential backoff and
// jitter, honoring a Retry-After header when the server supplies one,
// refreshing the OAuth token for at most one retry on a 401, and — for a
// modifying method (POST/PUT/PATCH/DELETE) — attaching and, if necessary,
// fetching and retrying with a CSRF token, since SAP's OData V2 services
// protect writes with both independently of one another. The caller owns
// req.Body: for retries to work with a body, req must have been built with
// GetBody set (as http.NewRequestWithContext does for common body types).
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	if needsCSRFToken(req.Method) {
		return c.doWriteWithCSRF(req)
	}

	return c.doAuthenticated(req)
}

// doAuthenticated performs the retry-with-backoff request and, on a 401,
// forces a fresh OAuth token and retries exactly once. It never itself
// handles CSRF; see doWriteWithCSRF for that layer.
func (c *Client) doAuthenticated(req *http.Request) (*http.Response, error) {
	resp, err := c.doWithRetries(req)
	if err != nil || c.invalidateToken == nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}

	// SAP can invalidate a token before its stated expiry, in which case the
	// cached token looks valid to us but is rejected downstream. Force a
	// fresh token and try exactly once more; if that also comes back 401,
	// the problem is not a stale token and we return it as-is.
	drainAndClose(resp.Body)
	c.invalidateToken()

	if req.GetBody != nil {
		body, bodyErr := req.GetBody()
		if bodyErr != nil {
			return nil, fmt.Errorf("http: rewinding request body for auth retry: %w", bodyErr)
		}
		req.Body = body
	}

	return c.doWithRetries(req)
}

// doWithRetries executes req, retrying transient failures with exponential
// backoff and jitter. It never itself distinguishes a 401; that is Do's job.
func (c *Client) doWithRetries(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("http: rewinding request body for retry: %w", err)
				}
				req.Body = body
			}
		}

		resp, err := c.transport.Do(req) //nolint:gosec // G704: req targets the SAP tenant host from this provider's own configuration (the "host" attribute or SAP_INTEGRATION_SUITE_HOST), never a value read from a request the provider itself receives — there is no untrusted caller who supplies this URL at runtime
		if err != nil {
			lastErr = err
			if !shouldRetryError(req.Context(), err) || attempt == c.maxRetries {
				return nil, err
			}
			if waitErr := c.wait(req.Context(), attempt, 0); waitErr != nil {
				return nil, waitErr
			}
			continue
		}

		if !isRetryable(resp.StatusCode) || attempt == c.maxRetries {
			return resp, nil
		}

		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		drainAndClose(resp.Body)

		if waitErr := c.wait(req.Context(), attempt, retryAfter); waitErr != nil {
			return nil, waitErr
		}
	}

	return nil, lastErr
}

func shouldRetryError(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	return err != nil
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 4096))
	_ = body.Close()
}

// wait blocks for the backoff duration for attempt, or until ctx is done. A
// non-zero retryAfter overrides the computed backoff, per RFC 7231.
func (c *Client) wait(ctx context.Context, attempt int, retryAfter time.Duration) error {
	delay := retryAfter
	if delay <= 0 {
		delay = backoffDelay(attempt, c.baseDelay, c.maxDelay)
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// backoffDelay computes an exponential backoff with full jitter, capped at
// maxDelay.
func backoffDelay(attempt int, base, maxDelay time.Duration) time.Duration {
	capped := math.Min(float64(maxDelay), float64(base)*math.Pow(2, float64(attempt)))
	if capped <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(capped)))
	if err != nil {
		return time.Duration(capped)
	}
	return time.Duration(n.Int64())
}

// parseRetryAfter parses a Retry-After header value expressed in seconds.
// SAP APIs consistently use the delay-seconds form rather than an HTTP-date;
// an unparsable or absent header yields zero, so the caller falls back to
// computed backoff.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// ReadLimited reads up to MaxResponseBytes from r and closes it, returning an
// error if the body was truncated.
func ReadLimited(r io.ReadCloser) ([]byte, error) {
	defer func() { _ = r.Close() }()

	limited := io.LimitReader(r, MaxResponseBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("http: reading response body: %w", err)
	}
	if len(data) > MaxResponseBytes {
		return nil, fmt.Errorf("http: response body exceeded %d bytes", MaxResponseBytes)
	}
	return data, nil
}
