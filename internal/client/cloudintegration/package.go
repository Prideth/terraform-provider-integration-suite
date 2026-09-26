// Package cloudintegration is the API client for SAP Cloud Integration's
// public "Integration Content" and "Security Content" OData V2 APIs
// (published as the CloudIntegrationAPI package on SAP Business Accelerator
// Hub). It knows the SAP wire format; it has no knowledge of Terraform.
package cloudintegration

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

// Package is the wire representation of an IntegrationPackages entity.
type Package struct {
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	Description  string `json:"Description,omitempty"`
	ShortText    string `json:"ShortText,omitempty"`
	Version      string `json:"Version,omitempty"`
	Vendor       string `json:"Vendor,omitempty"`
	Mode         string `json:"Mode,omitempty"`
	CreatedBy    string `json:"CreatedBy,omitempty"`
	CreationDate string `json:"CreationDate,omitempty"`
	ModifiedBy   string `json:"ModifiedBy,omitempty"`
	ModifiedDate string `json:"ModifiedDate,omitempty"`
}

const integrationPackagesEntitySet = "IntegrationPackages"

// GetPackage reads a single integration package by ID. A missing package
// surfaces as an *apierror.Error with IsNotFound() == true.
func (c *Client) GetPackage(ctx context.Context, id string) (*Package, error) {
	path := v2.BuildPath(integrationPackagesEntitySet, v2.KeyPredicate(id), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var pkg Package
	if err := v2.DecodeEntity(body, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// CreatePackage creates a new integration package. ShortText is required by
// SAP.
func (c *Client) CreatePackage(ctx context.Context, pkg Package) (*Package, error) {
	payload, err := json.Marshal(packageWriteRequest{ID: pkg.ID, Name: pkg.Name, Description: pkg.Description, ShortText: pkg.ShortText})
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding package: %w", err)
	}

	body, err := c.odata.Post(ctx, integrationPackagesEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created Package
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// packageWriteRequest is the body of create and update. A tenant test
// (September 2026) settled it: create without ShortText fails with
// "Property 'ShortText' cannot be empty", and {Id, Name, Description,
// ShortText} is accepted by both POST (201) and PUT (202).
type packageWriteRequest struct {
	ID          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	ShortText   string `json:"ShortText"`
}

// UpdatePackage changes an integration package's name, description and short
// text. The same tenant test showed that PATCH answers 501 Not Implemented
// and PUT answers 202 with the change read back, so this is a PUT. The
// package ID is immutable.
func (c *Client) UpdatePackage(ctx context.Context, id string, pkg Package) error {
	payload, err := json.Marshal(packageWriteRequest{ID: id, Name: pkg.Name, Description: pkg.Description, ShortText: pkg.ShortText})
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding package: %w", err)
	}

	path := v2.BuildPath(integrationPackagesEntitySet, v2.KeyPredicate(id), "")
	_, err = c.odata.Put(ctx, path, payload)
	return err
}

// PlainDescription undoes the HTML wrapping SAP applies to package
// descriptions: a tenant stored "tf-acc probe" and returned
// "<p>tf-acc probe</p>", and an empty description as "<p></p>". A single
// surrounding paragraph without further markup is removed and its entities
// decoded; any other value is returned unchanged.
func PlainDescription(s string) string {
	if !strings.HasPrefix(s, "<p>") || !strings.HasSuffix(s, "</p>") {
		return s
	}
	inner := s[len("<p>") : len(s)-len("</p>")]
	if strings.ContainsAny(inner, "<>") {
		return s
	}
	return html.UnescapeString(inner)
}

// DeletePackage deletes an integration package. A 404 is returned to the
// caller as an *apierror.Error so resource Delete implementations can treat
// it as already-deleted.
func (c *Client) DeletePackage(ctx context.Context, id string) error {
	path := v2.BuildPath(integrationPackagesEntitySet, v2.KeyPredicate(id), "")
	return c.odata.Delete(ctx, path)
}
