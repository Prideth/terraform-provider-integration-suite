package v2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	sapthttp "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/http"
)

// doer is the minimal interface the shared retrying HTTP client satisfies;
// tests can substitute their own implementation.
type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a small OData V2 request helper: it joins a base URL with an
// entity path, sets the headers SAP's OData V2 services expect, and turns
// non-2xx responses into a *apierror.Error via ParseError.
type Client struct {
	http    doer
	baseURL string
}

// New builds an OData V2 client rooted at baseURL (for example
// "https://tenant.example/api/v1"). http is typically a
// *sapthttp.Client wrapping an OAuth2-authenticated transport.
func New(httpClient doer, baseURL string) *Client {
	return &Client{http: httpClient, baseURL: baseURL}
}

// Get issues a GET request against path (an entity set, optionally with a
// key predicate and query string) and returns the raw response body on
// success.
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

// Post issues a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

// Put issues a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

// Delete issues a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}

func (c *Client) do(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	url := c.baseURL + "/" + path

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("odata: building request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req) //nolint:bodyclose // resp.Body is always closed inside sapthttp.ReadLimited below
	if err != nil {
		return nil, fmt.Errorf("odata: request failed: %w", err)
	}

	respBody, err := sapthttp.ReadLimited(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, ParseError(resp.StatusCode, respBody)
	}

	return respBody, nil
}
