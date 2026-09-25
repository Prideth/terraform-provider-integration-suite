package cloudintegration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const integrationAdapterDesigntimeArtifactsEntitySet = "IntegrationAdapterDesigntimeArtifacts"

// IntegrationAdapter is the wire representation of an
// IntegrationAdapterDesigntimeArtifacts entity: a custom adapter built with
// the SAP Adapter SDK and imported into a design-time package as a *.esa
// archive (Cloud Foundry only).
//
// The tenant $metadata defines exactly Id (the only key), Version, PackageId,
// Name, ArtifactContent (Edm.Binary) and Description. Earlier releases also
// sent Type and Application, UI classifications that are not properties of
// this entity; they were removed. SAP fills Name, Version and Description
// from the *.esa file itself.
type IntegrationAdapter struct {
	ID          string `json:"Id"`
	Name        string `json:"Name,omitempty"`
	PackageID   string `json:"PackageId,omitempty"`
	Version     string `json:"Version,omitempty"`
	Description string `json:"Description,omitempty"`

	// Content is the base64-encoded *.esa content, transported opaquely and
	// only set on Create; Read does not return it.
	Content string `json:"ArtifactContent,omitempty"`
}

// GetIntegrationAdapter reads a custom integration adapter's metadata by
// its tenant-wide-unique Id. SAP's documentation confirms Id uniqueness
// ("The integration adapter ID needs to be unique across the tenant"), and
// this client addresses the entity by Id alone, matching the confirmed
// Delete request shape (see DeleteIntegrationAdapter).
func (c *Client) GetIntegrationAdapter(ctx context.Context, id string) (*IntegrationAdapter, error) {
	path := v2.BuildPath(integrationAdapterDesigntimeArtifactsEntitySet, v2.KeyPredicate(id), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var adapter IntegrationAdapter
	if err := v2.DecodeEntity(body, &adapter); err != nil {
		return nil, err
	}
	return &adapter, nil
}

// CreateIntegrationAdapter imports a new custom integration adapter from
// *.esa content. content must already be the raw (not yet base64-encoded)
// *.esa bytes.
//
// SAP documents that importing an Id that already exists on the tenant is
// rejected as an error ("If there's already an integration adapter with
// the same ID, the system throws an error"), which is why this client has
// no UpdateIntegrationAdapter method: no public API for updating an
// existing adapter's content in place was confirmed, and this documented
// duplicate-ID error is positive evidence against one existing via a
// second Create call with the same Id. See
// sapintegrationsuite_integration_adapter's schema for the resulting
// RequiresReplace design.
func (c *Client) CreateIntegrationAdapter(ctx context.Context, adapter IntegrationAdapter, content []byte) (*IntegrationAdapter, error) {
	payload, err := json.Marshal(IntegrationAdapter{
		ID:        adapter.ID,
		Name:      adapter.Name,
		PackageID: adapter.PackageID,
		Content:   base64.StdEncoding.EncodeToString(content),
	})
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding integration adapter: %w", err)
	}

	body, err := c.odata.Post(ctx, integrationAdapterDesigntimeArtifactsEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created IntegrationAdapter
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteIntegrationAdapter deletes a custom integration adapter by Id.
// Confirmed directly from SAP's own "Integration Adapter Example
// Requests, Cloud Foundry Environment" documentation: `DELETE
// /api/v1/IntegrationAdapterDesigntimeArtifacts(Id='SubsystemSymbolicName1')`.
// Whether this can succeed while the adapter is still deployed, and
// whether it removes every version of the artifact or only the current
// one, is not confirmed by that same source — see
// sapintegrationsuite_integration_adapter's schema documentation.
func (c *Client) DeleteIntegrationAdapter(ctx context.Context, id string) error {
	path := v2.BuildPath(integrationAdapterDesigntimeArtifactsEntitySet, v2.KeyPredicate(id), "")
	return c.odata.Delete(ctx, path)
}

// DeployIntegrationAdapter triggers deployment of a custom integration
// adapter. Confirmed directly from SAP's own "Integration Adapter Example
// Requests, Cloud Foundry Environment" documentation: `POST
// /api/v1/DeployIntegrationAdapterDesigntimeArtifact?Id='SubsystemSymbolicName1'`.
//
// Two details are worth calling out because they differ from the sibling
// design-time artifact types' deploy actions elsewhere in this codebase:
//   - The action name is singular ("...Artifact"), the same convention
//     every other design-time artifact type's deploy action already
//     follows in this provider — despite this being easy to mis-assume as
//     plural ("...Artifacts") by loose analogy with the entity set's own
//     name, SAP's own example request confirms the singular form.
//   - There is no Version query parameter, unlike
//     DeployIntegrationDesigntimeArtifact/DeployValueMappingDesigntimeArtifact/
//     DeployScriptCollectionDesigntimeArtifact/DeployMessageMappingDesigntimeArtifact,
//     which all take both Id and Version. This is consistent with Id being
//     the adapter entity's only confirmed key (see GetIntegrationAdapter):
//     there is nothing else to disambiguate.
func (c *Client) DeployIntegrationAdapter(ctx context.Context, id string) error {
	path := fmt.Sprintf("DeployIntegrationAdapterDesigntimeArtifact?Id='%s'", v2.EscapeLiteral(id))
	_, err := c.odata.Post(ctx, path, nil)
	return err
}
