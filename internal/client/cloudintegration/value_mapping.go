package cloudintegration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const valueMappingDesigntimeArtifactsEntitySet = "ValueMappingDesigntimeArtifacts"

// ValueMapping is the wire representation of a ValueMappingDesigntimeArtifacts
// entity.
type ValueMapping struct {
	ID        string `json:"Id"`
	Name      string `json:"Name"`
	PackageID string `json:"PackageId"`
	Version   string `json:"Version,omitempty"`

	// Content is the base64-encoded value mapping project content (SAP's
	// own design-time export/import format for this artifact type). It is
	// only populated on Create/Update requests; Read does not return it.
	Content string `json:"ArtifactContent,omitempty"`
}

// GetValueMapping reads the active design-time version's metadata for the
// value mapping identified by mappingID within packageID.
func (c *Client) GetValueMapping(ctx context.Context, packageID, mappingID string) (*ValueMapping, error) {
	key, err := designtimeArtifactKey(mappingID, activeVersion)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Get(ctx, v2.BuildPath(valueMappingDesigntimeArtifactsEntitySet, key, ""))
	if err != nil {
		return nil, err
	}

	var mapping ValueMapping
	if err := v2.DecodeEntity(body, &mapping); err != nil {
		return nil, err
	}
	return &mapping, nil
}

// CreateValueMapping creates a new value mapping design-time artifact from
// uploaded content. content must already be the raw (not yet
// base64-encoded) bytes, and per SAP's own documented constraint must
// contain at least one mapping entry — a value mapping cannot be saved
// empty.
func (c *Client) CreateValueMapping(ctx context.Context, packageID, mappingID, name string, content []byte) (*ValueMapping, error) {
	payload, err := json.Marshal(ValueMapping{
		ID:        mappingID,
		Name:      name,
		PackageID: packageID,
		Content:   base64.StdEncoding.EncodeToString(content),
	})
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding value mapping: %w", err)
	}

	body, err := c.odata.Post(ctx, valueMappingDesigntimeArtifactsEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var mapping ValueMapping
	if err := v2.DecodeEntity(body, &mapping); err != nil {
		return nil, err
	}
	return &mapping, nil
}

// UpdateValueMapping uploads new content for an existing value mapping,
// creating a new design-time version under the same mapping ID, following
// the same PUT-against-the-keyed-entity convention already implemented and
// tested for sapintegrationsuite_integration_flow's sibling entity set.
//
// This is a documented assumption, not a confirmed fact: SAP separately
// documents a distinct ValueMappingDesigntimeArtifactSaveAsVersion action
// that takes a caller-supplied new version identifier, which may be the
// API's actual intended update path instead of (or in addition to) PUT. See
// docs/sap-api-references.md and docs/resource-design.md for the full
// reasoning and what would need to change if this assumption is wrong.
func (c *Client) UpdateValueMapping(ctx context.Context, mappingID, name string, content []byte) (*ValueMapping, error) {
	payload, err := json.Marshal(ValueMapping{
		Name:    name,
		Content: base64.StdEncoding.EncodeToString(content),
	})
	if err != nil {
		return nil, fmt.Errorf("cloudintegration: encoding value mapping: %w", err)
	}

	key, err := designtimeArtifactKey(mappingID, activeVersion)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Put(ctx, v2.BuildPath(valueMappingDesigntimeArtifactsEntitySet, key, ""), payload)
	if err != nil {
		return nil, err
	}

	var mapping ValueMapping
	if err := v2.DecodeEntity(body, &mapping); err != nil {
		return nil, err
	}
	return &mapping, nil
}

// DeleteValueMapping deletes a value mapping design-time artifact (all
// versions).
func (c *Client) DeleteValueMapping(ctx context.Context, mappingID string) error {
	key, err := designtimeArtifactKey(mappingID, activeVersion)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, v2.BuildPath(valueMappingDesigntimeArtifactsEntitySet, key, ""))
}

// DeployValueMapping triggers deployment of the given design-time version
// of a value mapping. It returns immediately once SAP has accepted the
// deployment request; callers must poll GetRuntimeArtifact for completion,
// since deployment is asynchronous.
func (c *Client) DeployValueMapping(ctx context.Context, mappingID, version string) error {
	path := fmt.Sprintf("DeployValueMappingDesigntimeArtifact?Id='%s'&Version='%s'",
		v2.EscapeLiteral(mappingID), v2.EscapeLiteral(version))
	_, err := c.odata.Post(ctx, path, nil)
	return err
}
