// Package apicomposition is the client for API Composition's Configuration
// API, which manages business data graphs. The contract follows SAP's
// "Configuration API Specification and Usage" and "Business Data Graph
// Configuration File" pages. The bodies are plain JSON without an OData V2
// "d" envelope, and the API is served from its own region-specific host
// (for example https://eu10.graph.sap) with credentials from an instance of
// the API Composition service, plan "configuration". See
// docs/guides/api-composition.md for what SAP documents and what is
// inferred.
package apicomposition

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	sapthttp "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/http"
)

// HTTPDoer is the transport this client needs: the shared retrying
// *http.Client from internal/client/http satisfies this directly.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the API Composition Configuration API client.
type Client struct {
	http    HTTPDoer
	baseURL string
}

// New builds an API Composition client. host is the region-specific host,
// for example https://eu10.graph.sap as in SAP's examples; the service path
// "/configuration/v1/sap.graph" is appended.
func New(httpClient HTTPDoer, host string) *Client {
	baseURL := strings.TrimRight(host, "/") + "/configuration/v1/sap.graph"
	return &Client{http: httpClient, baseURL: baseURL}
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) post(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) patch(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.do(ctx, http.MethodPatch, path, body)
}

func (c *Client) delete(ctx context.Context, path string) error {
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
		return nil, fmt.Errorf("apicomposition: building request: %w", err)
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
		return nil, fmt.Errorf("apicomposition: request failed: %w", err)
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
