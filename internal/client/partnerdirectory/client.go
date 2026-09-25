// Package partnerdirectory implements a client for SAP Integration Suite's
// Partner Directory OData V2 API: Partners (read-only discovery),
// StringParameters, BinaryParameters, AlternativePartners, AuthorizedUsers,
// and UserCredentialParameters. It shares the same "/api/v1" service root
// as the Integration Content and Security Content APIs in
// internal/client/cloudintegration, but is kept in its own package rather
// than folded into that one: Partner Directory is a distinct SAP product
// area with its own entity model, its own authorization role
// (AuthGroup_TenantPartnerDirectoryConfigurator), and — for
// UserCredentialParameters — security-sensitive semantics that do not
// belong mixed into a general-purpose Cloud Integration client file.
package partnerdirectory

import (
	"net/http"
	"strings"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

// HTTPDoer is the transport the Partner Directory client needs: the shared
// retrying *http.Client from internal/client/http satisfies this directly.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the Partner Directory API client.
type Client struct {
	odata *v2.Client
	host  string
}

// New builds a Partner Directory client. host is the tenant's Integration
// Suite base URL (for example
// https://tenant.it-cpi.cfapps.eu10.hana.ondemand.com); the API's "/api/v1"
// path is appended automatically, the same service root used for
// Integration Content and Security Content.
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
