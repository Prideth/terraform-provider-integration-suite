// Package cloudintegration is the API client for SAP Cloud Integration's
// public "Integration Content" and "Security Content" OData V2 APIs
// (published as the CloudIntegrationAPI package on SAP Business Accelerator
// Hub). It knows the SAP wire format; it has no knowledge of Terraform.
package cloudintegration

import (
	"context"
	"encoding/json"
	"fmt"

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

// CreatePackage creates a new integration package.
func (c *Client) CreatePackage(ctx context.Context, pkg Package) (*Package, error) {
	payload, err := json.Marshal(pkg)
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

// UpdatePackage updates the mutable fields of an existing integration
// package (SAP only supports updating a subset of fields; the package ID is
// immutable). This uses PATCH rather than PUT: Terraform's schema only ever
// supplies name/description, and a PUT's full-replace semantics would risk
// resetting fields the schema does not track (ShortText, Vendor, Mode, ...)
// to their defaults.
func (c *Client) UpdatePackage(ctx context.Context, id string, pkg Package) error {
	payload, err := json.Marshal(pkg)
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding package: %w", err)
	}

	path := v2.BuildPath(integrationPackagesEntitySet, v2.KeyPredicate(id), "")
	_, err = c.odata.Patch(ctx, path, payload)
	return err
}

// DeletePackage deletes an integration package. A 404 is returned to the
// caller as an *apierror.Error so resource Delete implementations can treat
// it as already-deleted.
func (c *Client) DeletePackage(ctx context.Context, id string) error {
	path := v2.BuildPath(integrationPackagesEntitySet, v2.KeyPredicate(id), "")
	return c.odata.Delete(ctx, path)
}
