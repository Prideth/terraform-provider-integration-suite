package http

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// csrfCache holds the most recently fetched CSRF token and its associated
// session cookie (if SAP's fetch response set one), shared across every
// request this Client makes. SAP's OData V2 services on Cloud Foundry/BTP
// CSRF-protect modifying requests independently of OAuth: OAuth proves who
// the caller is, CSRF proves the write was not forged, and the two are
// handled as separate concerns here (see Client.Do and doWriteWithCSRF).
type csrfCache struct {
	mu     sync.Mutex
	token  string
	cookie string
}

func (c *csrfCache) snapshot() (token, cookie string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.token, c.cookie
}

func (c *csrfCache) store(token, cookie string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	c.cookie = cookie
}

// csrfProtectedMethods are the HTTP methods SAP's OData V2 services require
// a valid X-CSRF-Token for. GET, HEAD, and OPTIONS are never protected —
// fetching a token is itself a GET, so protecting it here would deadlock.
func needsCSRFToken(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// isCSRFTokenRequired reports whether resp is SAP's standard signal that a
// modifying request was rejected for a missing or stale CSRF token: a 403
// response carrying the response header "X-CSRF-Token: Required". Any other
// 403 (an authorization failure unrelated to CSRF) is left untouched.
func isCSRFTokenRequired(resp *http.Response) bool {
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		return false
	}
	return strings.EqualFold(resp.Header.Get("X-CSRF-Token"), "Required")
}

// applyCSRFHeaders sets the token and, if SAP's fetch response set one, the
// session cookie tying that token to a server-side session.
func applyCSRFHeaders(req *http.Request, token, cookie string) {
	req.Header.Set("X-CSRF-Token", token)
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
}

// cookieHeaderFromResponse rebuilds a Cookie request-header value from every
// Set-Cookie response header on resp, keeping only name=value pairs (cookie
// attributes like Path or HttpOnly are meaningless on the request side).
// Returns "" if the response set no cookies, which is the common case for a
// purely bearer-token-authenticated API client with no server-side session.
func cookieHeaderFromResponse(resp *http.Response) string {
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return ""
	}
	parts := make([]string, 0, len(cookies))
	for _, c := range cookies {
		parts = append(parts, c.Name+"="+c.Value)
	}
	return strings.Join(parts, "; ")
}

// doWriteWithCSRF sends a CSRF-protected modifying request, attaching
// whatever token this client has cached. If SAP rejects it specifically for
// a missing/stale token (isCSRFTokenRequired), it fetches a fresh token with
// exactly one GET against the same URL, caches it, and retries the original
// request exactly once. Any other failure — including a second CSRF
// rejection after the refresh — is returned to the caller as-is; refreshing
// further would not help since a freshly issued token was already rejected.
func (c *Client) doWriteWithCSRF(req *http.Request) (*http.Response, error) {
	if token, cookie := c.csrf.snapshot(); token != "" {
		applyCSRFHeaders(req, token, cookie)
	}

	resp, err := c.doAuthenticated(req)
	if err != nil || !isCSRFTokenRequired(resp) {
		return resp, err
	}
	drainAndClose(resp.Body)

	token, cookie, fetchErr := c.fetchCSRFToken(req)
	if fetchErr != nil {
		return nil, fetchErr
	}
	c.csrf.store(token, cookie)

	if req.GetBody != nil {
		body, bodyErr := req.GetBody()
		if bodyErr != nil {
			return nil, fmt.Errorf("http: rewinding request body for CSRF retry: %w", bodyErr)
		}
		req.Body = body
	}
	applyCSRFHeaders(req, token, cookie)

	return c.doAuthenticated(req)
}

// fetchCSRFToken issues the GET SAP's CSRF protocol documents: the same URL
// the modifying request targets, with X-CSRF-Token: fetch. It goes through
// doAuthenticated so OAuth token refresh still applies to it, but never
// through doWriteWithCSRF itself, since a GET is never CSRF-protected.
func (c *Client) fetchCSRFToken(originalReq *http.Request) (token, cookie string, err error) {
	fetchReq, err := http.NewRequestWithContext(originalReq.Context(), http.MethodGet, originalReq.URL.String(), nil)
	if err != nil {
		return "", "", fmt.Errorf("http: building CSRF token fetch request: %w", err)
	}
	fetchReq.Header.Set("X-CSRF-Token", "fetch")
	if c.userAgent != "" {
		fetchReq.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.doAuthenticated(fetchReq)
	if err != nil {
		return "", "", fmt.Errorf("http: fetching CSRF token: %w", err)
	}
	defer drainAndClose(resp.Body)

	token = resp.Header.Get("X-CSRF-Token")
	if token == "" || strings.EqualFold(token, "Required") {
		return "", "", fmt.Errorf("http: SAP did not return a CSRF token from the fetch request (status %d)", resp.StatusCode)
	}

	return token, cookieHeaderFromResponse(resp), nil
}
