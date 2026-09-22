package cloudintegration

import (
	"context"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

// runtimeArtifactsEntitySet is the shared runtime-artifacts entity set: SAP
// documents its GET as the general "Runtime Status API" for "currently
// deployed integration artifacts", not one entity set per design-time
// artifact type. Integration flows, value mappings, and (per that same
// documentation) other deployable design-time artifact types are all read
// and undeployed through this single entity set.
const runtimeArtifactsEntitySet = "IntegrationRuntimeArtifacts"

// RuntimeArtifact is the wire representation of an IntegrationRuntimeArtifacts
// entity: the deployed state of a design-time artifact (integration flow,
// value mapping, ...).
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

// GetRuntimeArtifact reads the current runtime deployment status of a
// deployed design-time artifact (identified by its design-time ID, which is
// also the runtime artifact's ID). A missing deployment surfaces as a 404
// *apierror.Error.
func (c *Client) GetRuntimeArtifact(ctx context.Context, id string) (*RuntimeArtifact, error) {
	path := v2.BuildPath(runtimeArtifactsEntitySet, v2.KeyPredicate(id), "")

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

// UndeployRuntimeArtifact removes the runtime deployment of a design-time
// artifact. A 404 means it was already undeployed.
func (c *Client) UndeployRuntimeArtifact(ctx context.Context, id string) error {
	path := v2.BuildPath(runtimeArtifactsEntitySet, v2.KeyPredicate(id), "")
	return c.odata.Delete(ctx, path)
}
