package cloudintegration

import (
	"context"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

// activeVersion is the OData V2 key literal SAP's Integration Content API
// accepts to mean "the current active design-time version" without the
// caller needing to track version numbers itself. It is shared by every
// versioned design-time artifact type in this API family (integration
// flows, value mappings, ...).
const activeVersion = "active"

// designtimeArtifactKey builds the composite (Id, Version) key predicate
// shared by every versioned design-time artifact entity set in this API.
func designtimeArtifactKey(id, version string) (string, error) {
	return v2.CompositeKeyPredicate("Id", id, "Version", version)
}

// saveAsVersion calls one of the <Artifact>SaveAsVersion function imports
// (parameters Id and SaveAsVersion, POST, returning the saved artifact), as SAP
// documents for IntegrationDesigntimeArtifactSaveAsVersion: upload the content
// with PUT first, then save it under an explicit version.
func saveAsVersion[T any](ctx context.Context, c *Client, functionImport, id, version string) (*T, error) {
	path := fmt.Sprintf("%s?Id='%s'&SaveAsVersion='%s'", functionImport, v2.EscapeLiteral(id), v2.EscapeLiteral(version))
	body, err := c.odata.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var saved T
	if err := v2.DecodeEntity(body, &saved); err != nil {
		return nil, err
	}
	return &saved, nil
}

// SaveIntegrationFlowAsVersion saves the current content of an integration
// flow under an explicit version such as "1.0.3".
func (c *Client) SaveIntegrationFlowAsVersion(ctx context.Context, flowID, version string) (*IntegrationFlow, error) {
	return saveAsVersion[IntegrationFlow](ctx, c, "IntegrationDesigntimeArtifactSaveAsVersion", flowID, version)
}

// SaveMessageMappingAsVersion saves the current content of a message mapping
// under an explicit version.
func (c *Client) SaveMessageMappingAsVersion(ctx context.Context, mappingID, version string) (*MessageMapping, error) {
	return saveAsVersion[MessageMapping](ctx, c, "MessageMappingDesigntimeArtifactSaveAsVersion", mappingID, version)
}

// SaveScriptCollectionAsVersion saves the current content of a script
// collection under an explicit version.
func (c *Client) SaveScriptCollectionAsVersion(ctx context.Context, scriptCollectionID, version string) (*ScriptCollection, error) {
	return saveAsVersion[ScriptCollection](ctx, c, "ScriptCollectionDesigntimeArtifactSaveAsVersion", scriptCollectionID, version)
}
