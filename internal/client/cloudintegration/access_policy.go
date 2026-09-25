package cloudintegration

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

// The wire contract below follows SAP's own access-policy automation in
// github.com/SAP/cicd-actions-for-sap-integration-suite (download-, upload-
// and delete-access-policy actions): both entity sets are keyed by Edm.Int64,
// references are created and deleted through the top-level ArtifactReferences
// set, and a reference is bound to its policy through the AccessPolicy
// navigation property in the create payload.
const (
	accessPoliciesEntitySet      = "AccessPolicies"
	artifactReferencesEntitySet  = "ArtifactReferences"
	accessPolicyReferencesNavKey = "ArtifactReferences"
)

// AccessPolicy is the wire representation of an AccessPolicies entity. Id is
// an Edm.Int64 that OData V2 JSON serializes as a string.
type AccessPolicy struct {
	ID          string `json:"Id,omitempty"`
	RoleName    string `json:"RoleName"`
	Description string `json:"Description"`
}

func accessPolicyPath(id string) (string, error) {
	key, err := v2.Int64KeyPredicate(id)
	if err != nil {
		return "", fmt.Errorf("cloudintegration: access policy ID: %w", err)
	}
	return accessPoliciesEntitySet + key, nil
}

// GetAccessPolicy reads a single access policy by its SAP-assigned ID.
func (c *Client) GetAccessPolicy(ctx context.Context, id string) (*AccessPolicy, error) {
	path, err := accessPolicyPath(id)
	if err != nil {
		return nil, err
	}

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

// FindAccessPolicyByRoleName returns the access policy whose RoleName equals
// roleName, or nil if none exists. RoleName is unique per tenant and is the
// identifier SAP's own tooling uses to locate a policy across environments.
func (c *Client) FindAccessPolicyByRoleName(ctx context.Context, roleName string) (*AccessPolicy, error) {
	query := v2.Query{Filter: v2.FilterEquals("RoleName", roleName)}.Encode()

	body, err := c.odata.Get(ctx, v2.BuildPath(accessPoliciesEntitySet, "", query))
	if err != nil {
		return nil, err
	}

	var policies []AccessPolicy
	if err := v2.DecodeCollection(body, &policies); err != nil {
		return nil, err
	}
	for i := range policies {
		if policies[i].RoleName == roleName {
			return &policies[i], nil
		}
	}
	return nil, nil
}

// CreateAccessPolicy creates a new access policy.
func (c *Client) CreateAccessPolicy(ctx context.Context, policy AccessPolicy) (*AccessPolicy, error) {
	payload, err := json.Marshal(AccessPolicy{RoleName: policy.RoleName, Description: policy.Description})
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

// UpdateAccessPolicy replaces the policy's writable fields with a PUT of
// RoleName and Description, the same payload SAP's own upload action sends.
// OData V2 PUT replaces the whole entity, so RoleName is always resent
// unchanged rather than omitted.
func (c *Client) UpdateAccessPolicy(ctx context.Context, id string, policy AccessPolicy) error {
	path, err := accessPolicyPath(id)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(AccessPolicy{RoleName: policy.RoleName, Description: policy.Description})
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding access policy: %w", err)
	}

	_, err = c.odata.Put(ctx, path, payload)
	return err
}

// DeleteAccessPolicy deletes an access policy. SAP removes the policy's
// artifact references together with it.
func (c *Client) DeleteAccessPolicy(ctx context.Context, id string) error {
	path, err := accessPolicyPath(id)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, path)
}

// AccessPolicyReference is the wire representation of an ArtifactReferences
// entity: one match rule of an access policy.
type AccessPolicyReference struct {
	ID                 string `json:"Id,omitempty"`
	Name               string `json:"Name"`
	Description        string `json:"Description"`
	Type               string `json:"Type"`
	ConditionAttribute string `json:"ConditionAttribute"`
	ConditionType      string `json:"ConditionType"`
	ConditionValue     string `json:"ConditionValue"`
}

type accessPolicyLink struct {
	ID string `json:"Id"`
}

type accessPolicyReferenceCreate struct {
	AccessPolicyReference
	AccessPolicy accessPolicyLink `json:"AccessPolicy"`
}

func artifactReferencePath(id string) (string, error) {
	key, err := v2.Int64KeyPredicate(id)
	if err != nil {
		return "", fmt.Errorf("cloudintegration: artifact reference ID: %w", err)
	}
	return artifactReferencesEntitySet + key, nil
}

// FindAccessPolicyReference returns the reference with ID referenceID from
// the given policy's reference list, or nil if the policy exists but no longer
// holds it. Reading through the policy, rather than ArtifactReferences(<id>L)
// directly, also proves the reference belongs to that policy.
func (c *Client) FindAccessPolicyReference(ctx context.Context, policyID, referenceID string) (*AccessPolicyReference, error) {
	refs, err := c.ListAccessPolicyReferences(ctx, policyID)
	if err != nil {
		return nil, err
	}
	for i := range refs {
		if refs[i].ID == referenceID {
			return &refs[i], nil
		}
	}
	return nil, nil
}

// ListAccessPolicyReferences returns every artifact reference of a policy
// through its ArtifactReferences navigation property.
func (c *Client) ListAccessPolicyReferences(ctx context.Context, policyID string) ([]AccessPolicyReference, error) {
	path, err := accessPolicyPath(policyID)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Get(ctx, path+"/"+accessPolicyReferencesNavKey)
	if err != nil {
		return nil, err
	}

	var refs []AccessPolicyReference
	if err := v2.DecodeCollection(body, &refs); err != nil {
		return nil, err
	}
	return refs, nil
}

// CreateAccessPolicyReference adds a new artifact reference to the policy
// identified by policyID.
func (c *Client) CreateAccessPolicyReference(ctx context.Context, policyID string, ref AccessPolicyReference) (*AccessPolicyReference, error) {
	if _, err := v2.Int64KeyPredicate(policyID); err != nil {
		return nil, fmt.Errorf("cloudintegration: access policy ID: %w", err)
	}

	ref.ID = ""
	payload, err := json.Marshal(accessPolicyReferenceCreate{
		AccessPolicyReference: ref,
		AccessPolicy:          accessPolicyLink{ID: policyID},
	})
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding access policy reference: %w", err)
	}

	body, err := c.odata.Post(ctx, artifactReferencesEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created AccessPolicyReference
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteAccessPolicyReference removes a single artifact reference.
func (c *Client) DeleteAccessPolicyReference(ctx context.Context, referenceID string) error {
	path, err := artifactReferencePath(referenceID)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, path)
}
