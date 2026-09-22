package cloudintegration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const (
	designtimeArtifactsEntitySet = "IntegrationDesigntimeArtifacts"
	runtimeArtifactsEntitySet    = "IntegrationRuntimeArtifacts"

	// activeVersion is the OData V2 key literal SAP's Integration Content
	// API accepts to mean "the current active design-time version" without
	// the caller needing to track version numbers itself.
	activeVersion = "active"
)

// IntegrationFlow is the wire representation of an
// IntegrationDesigntimeArtifacts entity.
type IntegrationFlow struct {
	ID        string `json:"Id"`
	Name      string `json:"Name"`
	PackageID string `json:"PackageId"`
	Version   string `json:"Version,omitempty"`

	// Content is the base64-encoded ZIP of the integration flow project. It
	// is only populated on Create/Update requests; Read does not return it
	// (SAP does not include binary content in the metadata response).
	Content string `json:"ArtifactContent,omitempty"`
}

// integrationFlowKey builds the composite (Id, Version) key predicate used
// by the design-time artifact entity set.
func integrationFlowKey(id, version string) (string, error) {
	return v2.CompositeKeyPredicate("Id", id, "Version", version)
}

// GetIntegrationFlow reads the active design-time version's metadata for
// the flow identified by flowID within packageID.
func (c *Client) GetIntegrationFlow(ctx context.Context, packageID, flowID string) (*IntegrationFlow, error) {
	key, err := integrationFlowKey(flowID, activeVersion)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Get(ctx, v2.BuildPath(designtimeArtifactsEntitySet, key, ""))
	if err != nil {
		return nil, err
	}

	var flow IntegrationFlow
	if err := v2.DecodeEntity(body, &flow); err != nil {
		return nil, err
	}
	return &flow, nil
}

// CreateIntegrationFlow creates a new integration flow design-time artifact
// from a ZIP project archive. content must already be the raw (not yet
// base64-encoded) ZIP bytes.
func (c *Client) CreateIntegrationFlow(ctx context.Context, packageID, flowID, name string, content []byte) (*IntegrationFlow, error) {
	payload, err := json.Marshal(IntegrationFlow{
		ID:        flowID,
		Name:      name,
		PackageID: packageID,
		Content:   base64.StdEncoding.EncodeToString(content),
	})
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding integration flow: %w", err)
	}

	body, err := c.odata.Post(ctx, designtimeArtifactsEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var flow IntegrationFlow
	if err := v2.DecodeEntity(body, &flow); err != nil {
		return nil, err
	}
	return &flow, nil
}

// UpdateIntegrationFlow uploads new content for an existing integration
// flow, creating a new design-time version under the same flow ID (SAP's
// design-time API is version-based, not in-place, which is why the flow's
// identity — packageID/flowID — never changes on update). Unlike creating a
// brand new flow, this targets the existing (Id, Version) entity with PUT
// rather than POSTing to the collection again, which is how OData V2
// distinguishes "create a new entity" from "update this one".
func (c *Client) UpdateIntegrationFlow(ctx context.Context, flowID, name string, content []byte) (*IntegrationFlow, error) {
	payload, err := json.Marshal(IntegrationFlow{
		Name:    name,
		Content: base64.StdEncoding.EncodeToString(content),
	})
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding integration flow: %w", err)
	}

	key, err := integrationFlowKey(flowID, activeVersion)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Put(ctx, v2.BuildPath(designtimeArtifactsEntitySet, key, ""), payload)
	if err != nil {
		return nil, err
	}

	var flow IntegrationFlow
	if err := v2.DecodeEntity(body, &flow); err != nil {
		return nil, err
	}
	return &flow, nil
}

// DeleteIntegrationFlow deletes an integration flow design-time artifact
// (all versions).
func (c *Client) DeleteIntegrationFlow(ctx context.Context, flowID string) error {
	key, err := integrationFlowKey(flowID, activeVersion)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, v2.BuildPath(designtimeArtifactsEntitySet, key, ""))
}

// RuntimeArtifact is the wire representation of an IntegrationRuntimeArtifacts
// entity: the deployed state of an integration flow.
//
// ErrorInfo is best-effort: SAP's monitoring UI sources failure detail from
// a separate error-information endpoint rather than always inlining it on
// this entity, so ErrorInfo may be empty even when Status is StatusError.
// Callers must not assume it is populated; this needs verification against
// a live tenant before being relied on for anything beyond a best-effort
// diagnostic message.
type RuntimeArtifact struct {
	ID         string `json:"Id"`
	Version    string `json:"Version"`
	Status     string `json:"Status"`
	DeployedBy string `json:"DeployedBy,omitempty"`
	DeployedOn string `json:"DeployedOn,omitempty"`
	ErrorInfo  string `json:"ErrorInformation,omitempty"`
}

// Runtime deployment status values reported by IntegrationRuntimeArtifacts.
const (
	StatusStarted  = "STARTED"
	StatusStarting = "STARTING"
	StatusStopped  = "STOPPED"
	StatusError    = "ERROR"
)

// DeployIntegrationFlow triggers deployment of the given design-time
// version of an integration flow. It returns immediately once SAP has
// accepted the deployment request; callers must poll GetRuntimeArtifact for
// completion, since deployment is asynchronous.
//
// version identifies which design-time version to deploy. Passing the
// literal "active" defers to whichever version is currently active, but the
// Terraform resource always passes the concrete version it read from the
// design-time artifact, so that changing the deployed version is a visible,
// plannable change rather than an implicit side effect of "whatever is
// active right now".
func (c *Client) DeployIntegrationFlow(ctx context.Context, flowID, version string) error {
	path := fmt.Sprintf("DeployIntegrationDesigntimeArtifact?Id='%s'&Version='%s'",
		v2.EscapeLiteral(flowID), v2.EscapeLiteral(version))
	_, err := c.odata.Post(ctx, path, nil)
	return err
}

// GetRuntimeArtifact reads the current runtime deployment status of an
// integration flow. A missing deployment surfaces as a 404 *apierror.Error.
func (c *Client) GetRuntimeArtifact(ctx context.Context, flowID string) (*RuntimeArtifact, error) {
	path := v2.BuildPath(runtimeArtifactsEntitySet, v2.KeyPredicate(flowID), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var artifact RuntimeArtifact
	if err := v2.DecodeEntity(body, &artifact); err != nil {
		return nil, err
	}
	return &artifact, nil
}

// UndeployIntegrationFlow removes the runtime deployment of an integration
// flow. A 404 means it was already undeployed.
func (c *Client) UndeployIntegrationFlow(ctx context.Context, flowID string) error {
	path := v2.BuildPath(runtimeArtifactsEntitySet, v2.KeyPredicate(flowID), "")
	return c.odata.Delete(ctx, path)
}
