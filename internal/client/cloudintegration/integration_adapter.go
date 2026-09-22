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
// IntegrationAdapterDesigntimeArtifacts entity: a custom Integration
// Adapter built with the SAP Adapter SDK and imported into a design-time
// package as a *.esa artifact. This is a Cloud Foundry-only artifact type
// per SAP's own documentation ("This information is relevant only when you
// use SAP Cloud Integration in the Cloud Foundry environment", repeated on
// every page describing it).
//
// Confidence per field, since SAP's public documentation for this entity
// is far sparser than for the sibling design-time artifact types in this
// same API (Integration Flow, Value Mapping, Message Mapping, Script
// Collection all have a dedicated, complete "Example Requests" page
// showing every operation; the equivalent adapter page shows only Deploy
// and Delete):
//   - ID: confirmed — SAP's own "Integration Adapter Example Requests, Cloud
//     Foundry Environment" page addresses an entity by
//     IntegrationAdapterDesigntimeArtifacts(Id='...') alone, not the
//     composite (Id, Version) key every other design-time artifact entity
//     in this API uses. This provider treats Id as the entity's sole key.
//   - PackageId: corroborated, not primary-confirmed for this entity
//     specifically — SAP's UI documentation confirms every custom adapter
//     belongs to an integration package at design time, and every sibling
//     design-time artifact type's confirmed Create payload includes
//     PackageId, but no adapter-specific Create example was found to
//     confirm the property name directly for this entity.
//   - Name, Version: SAP's UI documentation states plainly that these are
//     "auto filled" from the imported *.esa file, confirming they exist as
//     metadata, but not confirming their exact OData property names or
//     whether the API accepts (or ignores) a caller-supplied value on
//     Create.
//   - Type, Application: confirmed as UI concepts (a line-of-business
//     classification and a target-application classification,
//     respectively, both selected from a list at import time), not
//     confirmed as literal OData property names, and not confirmed to be
//     either free-form or a closed API-level enum — see the Type/Application
//     doc comment on the resource schema for why this provider does not
//     validate either as an enum.
//
// ArtifactContent (the base64-encoded *.esa content) is corroborated as
// the same field name every sibling design-time artifact type in this API
// uses for its content payload, not independently confirmed by an
// adapter-specific example request.
type IntegrationAdapter struct {
	ID          string `json:"Id"`
	Name        string `json:"Name,omitempty"`
	PackageID   string `json:"PackageId,omitempty"`
	Version     string `json:"Version,omitempty"`
	Type        string `json:"Type,omitempty"`
	Application string `json:"Application,omitempty"`

	// Content is the base64-encoded *.esa content. This client transports
	// it opaquely — it is never unpacked, executed, or inspected beyond
	// what the caller already validated (size bound, content hash) before
	// handing it to this client. It is only populated on Create requests;
	// Read does not return it.
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
		ID:          adapter.ID,
		Name:        adapter.Name,
		PackageID:   adapter.PackageID,
		Type:        adapter.Type,
		Application: adapter.Application,
		Content:     base64.StdEncoding.EncodeToString(content),
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
