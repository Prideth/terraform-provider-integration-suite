// Package securitycontent is the client for SAP Cloud Integration's public
// "Security Content" OData V2 API: the service that manages security
// material artifacts (user credentials, OAuth2 client credentials, keystore
// entries, and related artifact types) on an Integration Suite tenant. It
// is kept separate from internal/client/cloudintegration, even though both
// services share the same "/api/v1" host and OData v2 transport, because
// Security Content artifacts have fundamentally different secret semantics
// (write-only, never read back) than ordinary design-time content, and
// grouping them under one client package would blur that distinction for
// every future contributor reading this code.
package securitycontent

import (
	"net/http"
	"strings"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

// HTTPDoer is the transport the Security Content client needs: the shared
// retrying *http.Client from internal/client/http satisfies this directly.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the Security Content API client.
type Client struct {
	odata *v2.Client
	host  string
}

// New builds a Security Content client. host is the tenant's Integration
// Suite base URL (for example
// https://tenant.it-cpi.cfapps.eu10.hana.ondemand.com); the API's "/api/v1"
// path is appended automatically.
func New(httpClient HTTPDoer, host string) *Client {
	baseURL := strings.TrimRight(host, "/") + "/api/v1"
	return &Client{odata: v2.New(httpClient, baseURL), host: host}
}

// AtLocation returns a client for the same tenant that addresses the runtime
// with the given runtime location ID, such as an Edge Integration Cell, through
// /location/<id>/api/v1. An empty ID returns c unchanged (the cloud runtime).
func (c *Client) AtLocation(runtimeLocationID string) (*Client, error) {
	if runtimeLocationID == "" {
		return c, nil
	}
	root, err := v2.ServiceRoot(c.host, runtimeLocationID)
	if err != nil {
		return nil, err
	}
	return &Client{odata: c.odata.WithBaseURL(root), host: c.host}, nil
}
