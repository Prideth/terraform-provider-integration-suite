package apimanagementclassic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const apiProvidersEntitySet = "APIProviders"

// APIProvider is the wire representation of an APIProviders entity,
// restricted to the "Internet" connection type — the only one of the four
// documented connection types (Internet, On Premise, Open Connectors, Cloud
// Integration) with a confirmed field-level JSON mapping. See
// docs/sap-api-references.md for the evidence and the scope of this
// limitation.
//
// SAP's official documentation confirms no update operation exists for API
// providers created through this API (only Create, Read, and Delete); this
// provider's Terraform resource matches that by treating every field as
// RequiresReplace.
type APIProvider struct {
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// DestType selects the connection type. Only "INTERNET" is supported by
	// this client — see the package doc comment.
	DestType string `json:"destType"`

	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	UseSSL   bool   `json:"useSSL"`
	TrustAll bool   `json:"trustAll"`

	// PathPrefix and URL configure the optional Catalog Service Settings
	// (SAP Gateway service discovery); both are omitted from the request
	// entirely when empty, since they are documented as relevant only to
	// SAP Gateway-enabled systems.
	PathPrefix string `json:"pathPrefix,omitempty"`
	URL        string `json:"url,omitempty"`

	// AuthType selects the authentication method for connection requests
	// (confirmed value: "BASIC"). UserName/Password are only meaningful
	// when AuthType requires them.
	AuthType string `json:"authType,omitempty"`
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
}

// CreateAPIProvider creates a new API provider, then polls GET until the
// provider is visible (or ctx's deadline is reached) to defend against the
// documented up-to-~20-second eventual-consistency window between a
// modifying call and subsequent GET requests. The entity returned is the
// one echoed back by the POST response itself, not a re-fetched copy —
// SAP's Create response already contains the full created entity.
func (c *Client) CreateAPIProvider(ctx context.Context, provider APIProvider) (*APIProvider, error) {
	payload, err := json.Marshal(provider) //nolint:gosec // G117: this deliberately marshals the backend Basic-auth password into the request body sent to SAP's Create API -- that is the whole purpose of this call, not a leak. The password is supplied via the resource's write-only password_wo attribute (see resource_api_provider.go) and is never populated from a GET response, so it never reaches Terraform state or a log line.
	if err != nil {
		return nil, fmt.Errorf("apimanagementclassic: encoding API provider: %w", err)
	}

	body, err := c.odata.Post(ctx, apiProvidersEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created APIProvider
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}

	if err := pollUntilVisible(ctx, func(ctx context.Context) (bool, error) {
		_, err := c.GetAPIProvider(ctx, created.Name)
		if err == nil {
			return true, nil
		}
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return false, nil
		}
		return false, err
	}); err != nil {
		return nil, fmt.Errorf("apimanagementclassic: API provider %q was created but did not become visible: %w", created.Name, err)
	}

	return &created, nil
}

// GetAPIProvider reads a single API provider by name.
func (c *Client) GetAPIProvider(ctx context.Context, name string) (*APIProvider, error) {
	path := v2.BuildPath(apiProvidersEntitySet, v2.KeyPredicate(name), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var provider APIProvider
	if err := v2.DecodeEntity(body, &provider); err != nil {
		return nil, err
	}
	return &provider, nil
}

// ListAPIProviders reads every API provider visible to the caller.
func (c *Client) ListAPIProviders(ctx context.Context) ([]APIProvider, error) {
	body, err := c.odata.Get(ctx, apiProvidersEntitySet)
	if err != nil {
		return nil, err
	}

	var providers []APIProvider
	if err := v2.DecodeCollection(body, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

// DeleteAPIProvider deletes an API provider by name.
func (c *Client) DeleteAPIProvider(ctx context.Context, name string) error {
	path := v2.BuildPath(apiProvidersEntitySet, v2.KeyPredicate(name), "")
	return c.odata.Delete(ctx, path)
}
