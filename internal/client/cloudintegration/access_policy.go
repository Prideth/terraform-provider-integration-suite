package cloudintegration

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const (
	accessPoliciesEntitySet       = "AccessPolicies"
	accessPolicyArtifactsRelation = "ArtifactReferences"
)

// AccessPolicy is the wire representation of an AccessPolicies entity
// (Security Content API).
type AccessPolicy struct {
	ID          string `json:"Id,omitempty"`
	RoleName    string `json:"RoleName"`
	Description string `json:"Description,omitempty"`

	// ReconciliationStatus reflects replication of the policy to any
	// associated runtime (Integration Cell / Edge Integration Cell), where
	// the API reports one. It is computed, never set by the caller.
	ReconciliationStatus string `json:"ReconciliationStatus,omitempty"`
}

// GetAccessPolicy reads a single access policy by its SAP-assigned ID.
func (c *Client) GetAccessPolicy(ctx context.Context, id string) (*AccessPolicy, error) {
	path := v2.BuildPath(accessPoliciesEntitySet, v2.KeyPredicate(id), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var policy AccessPolicy
	if err := v2.DecodeEntity(body, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// CreateAccessPolicy creates a new access policy.
func (c *Client) CreateAccessPolicy(ctx context.Context, policy AccessPolicy) (*AccessPolicy, error) {
	payload, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding access policy: %w", err)
	}

	body, err := c.odata.Post(ctx, accessPoliciesEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created AccessPolicy
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateAccessPolicy updates an access policy's mutable fields (currently:
// description; the role name identifies the policy and is treated as
// immutable by the Terraform resource). This uses PATCH rather than PUT so
// that fields outside the Terraform schema are left untouched.
func (c *Client) UpdateAccessPolicy(ctx context.Context, id string, policy AccessPolicy) error {
	payload, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding access policy: %w", err)
	}

	path := v2.BuildPath(accessPoliciesEntitySet, v2.KeyPredicate(id), "")
	_, err = c.odata.Patch(ctx, path, payload)
	return err
}

// DeleteAccessPolicy deletes an access policy.
func (c *Client) DeleteAccessPolicy(ctx context.Context, id string) error {
	path := v2.BuildPath(accessPoliciesEntitySet, v2.KeyPredicate(id), "")
	return c.odata.Delete(ctx, path)
}

// AccessPolicyReference is the wire representation of a single artifact
// reference nested under an access policy.
type AccessPolicyReference struct {
	ID           string `json:"Id,omitempty"`
	ArtifactType string `json:"ArtifactType"`
	Attribute    string `json:"Attribute"`
	Operator     string `json:"Operator"`
	Value        string `json:"Value"`
}

// SupportedArtifactTypes lists the artifact types SAP currently documents as
// supported by access policy artifact references. The Terraform validator
// only accepts values from this list; it is intentionally not extended with
// any value that has not been confirmed in SAP's own documentation.
var SupportedArtifactTypes = []string{
	"IntegrationFlow",
	"ODataAPI",
	"RestAPI",
	"SoapAPI",
	"ScriptCollection",
	"ValueMapping",
	"MessageMapping",
	"MessageQueue",
	"GlobalDataStore",
	"GlobalVariable",
}

func accessPolicyReferencesPath(policyID string) string {
	return fmt.Sprintf("%s%s/%s", accessPoliciesEntitySet, v2.KeyPredicate(policyID), accessPolicyArtifactsRelation)
}

// GetAccessPolicyReference reads a single artifact reference by its
// composite (policy ID, reference ID) identity.
func (c *Client) GetAccessPolicyReference(ctx context.Context, policyID, referenceID string) (*AccessPolicyReference, error) {
	path := accessPolicyReferencesPath(policyID) + v2.KeyPredicate(referenceID)

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var ref AccessPolicyReference
	if err := v2.DecodeEntity(body, &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}

// CreateAccessPolicyReference adds a new artifact reference to an access
// policy.
func (c *Client) CreateAccessPolicyReference(ctx context.Context, policyID string, ref AccessPolicyReference) (*AccessPolicyReference, error) {
	payload, err := json.Marshal(ref)
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding access policy reference: %w", err)
	}

	body, err := c.odata.Post(ctx, accessPolicyReferencesPath(policyID), payload)
	if err != nil {
		return nil, err
	}

	var created AccessPolicyReference
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteAccessPolicyReference removes a single artifact reference from an
// access policy.
func (c *Client) DeleteAccessPolicyReference(ctx context.Context, policyID, referenceID string) error {
	path := accessPolicyReferencesPath(policyID) + v2.KeyPredicate(referenceID)
	return c.odata.Delete(ctx, path)
}
