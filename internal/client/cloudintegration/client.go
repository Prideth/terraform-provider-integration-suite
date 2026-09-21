package cloudintegration

import (
	"net/http"
	"strings"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

// HTTPDoer is the transport the Cloud Integration client needs: the shared
// retrying *http.Client from internal/client/http satisfies this directly.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the Cloud Integration API client, covering the "Integration
// Content" and "Security Content" OData V2 services under /api/v1.
type Client struct {
	odata *v2.Client
}

// New builds a Cloud Integration client. host is the tenant's Integration
// Suite base URL (for example
// https://tenant.it-cpi.cfapps.eu10.hana.ondemand.com); the API's "/api/v1"
// path is appended automatically.
func New(httpClient HTTPDoer, host string) *Client {
	baseURL := strings.TrimRight(host, "/") + "/api/v1"
	return &Client{odata: v2.New(httpClient, baseURL)}
}
