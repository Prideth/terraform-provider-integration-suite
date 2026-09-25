package cloudintegration

import (
	"testing"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/testutil/edmx"
)

// TestWireContractAgainstMetadata checks every wire struct, entity set, key
// and function import this package uses against a tenant $metadata document.
// It skips when no document is available (see internal/testutil/edmx).
func TestWireContractAgainstMetadata(t *testing.T) {
	m := edmx.Load(t)

	structs := []struct {
		entitySet string
		value     any
	}{
		{accessPoliciesEntitySet, AccessPolicy{}},
		{accessPoliciesEntitySet, accessPolicyLink{}},
		{artifactReferencesEntitySet, AccessPolicyReference{}},
		{artifactReferencesEntitySet, accessPolicyReferenceCreate{}},
		{customTagConfigurationsEntitySet, customTagConfigurationWriteRequest{}},
		{integrationAdapterDesigntimeArtifactsEntitySet, IntegrationAdapter{}},
		{integrationDesigntimeArtifactsEntitySet, IntegrationFlow{}},
		{messageMappingDesigntimeArtifactsEntitySet, MessageMapping{}},
		{numberRangesEntitySet, numberRangeWireModel{}},
		{integrationPackagesEntitySet, Package{}},
		{runtimeArtifactsEntitySet, RuntimeArtifact{}},
		{scriptCollectionDesigntimeArtifactsEntitySet, ScriptCollection{}},
		{serviceEndpointsEntitySet, ServiceEndpoint{}},
		{"EntryPoints", EntryPoint{}},
		{"APIDefinitions", APIDefinition{}},
		{valueMappingDesigntimeArtifactsEntitySet, ValueMapping{}},
	}
	for _, s := range structs {
		m.AssertStruct(t, s.entitySet, s.value)
	}

	m.AssertKey(t, accessPoliciesEntitySet, "Id", "Edm.Int64")
	m.AssertKey(t, artifactReferencesEntitySet, "Id", "Edm.Int64")
	m.AssertKey(t, integrationPackagesEntitySet, "Id", "Edm.String")
	m.AssertKey(t, integrationDesigntimeArtifactsEntitySet, "Id", "Edm.String", "Version", "Edm.String")
	m.AssertKey(t, messageMappingDesigntimeArtifactsEntitySet, "Id", "Edm.String", "Version", "Edm.String")
	m.AssertKey(t, scriptCollectionDesigntimeArtifactsEntitySet, "Id", "Edm.String", "Version", "Edm.String")
	m.AssertKey(t, valueMappingDesigntimeArtifactsEntitySet, "Id", "Edm.String", "Version", "Edm.String")
	m.AssertKey(t, integrationAdapterDesigntimeArtifactsEntitySet, "Id", "Edm.String")
	m.AssertKey(t, runtimeArtifactsEntitySet, "Id", "Edm.String")
	m.AssertKey(t, numberRangesEntitySet, "Name", "Edm.String")
	m.AssertKey(t, customTagConfigurationsEntitySet, "Id", "Edm.String")

	m.AssertFunctionImport(t, "DeployIntegrationDesigntimeArtifact", "POST", "Id", "Version")
	m.AssertFunctionImport(t, "DeployMessageMappingDesigntimeArtifact", "POST", "Id", "Version")
	m.AssertFunctionImport(t, "DeployScriptCollectionDesigntimeArtifact", "POST", "Id", "Version")
	m.AssertFunctionImport(t, "DeployValueMappingDesigntimeArtifact", "POST", "Id", "Version")
	m.AssertFunctionImport(t, "DeployIntegrationAdapterDesigntimeArtifact", "POST", "Id")
}
