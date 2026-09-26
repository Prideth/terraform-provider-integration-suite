package apimanagementclassic

import (
	"testing"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/testutil/edmx"
)

// TestWireContractAgainstMetadata checks every wire struct and entity set
// this package uses against the API portal's Management.svc $metadata. It
// skips when no document is available (see internal/testutil/edmx).
func TestWireContractAgainstMetadata(t *testing.T) {
	m := edmx.LoadFrom(t, "SAP_API_PORTAL_METADATA_FILE", ".specs/apim-management-metadata.xml")

	structs := []struct {
		entitySet string
		value     any
	}{
		{apiProvidersEntitySet, APIProvider{}},
		{apiProductsEntitySet, APIProduct{}},
		{apiProductsEntitySet, apiProductWire{}},
		{apiProductsEntitySet, apiProductReadWire{}},
		{apiProductAdditionalPropertiesEntity, APIProductAdditionalProperty{}},
		{certificateStoreReferencesEntitySet, CertificateStoreReference{}},
		{genericKeyMapEntriesEntitySet, KeyValueMap{}},
		{genericKeyMapEntriesEntitySet, keyValueMapWire{}},
		{"GenericKeyMapEntryValues", keyValueMapEntryValueWire{}},
	}
	for _, s := range structs {
		m.AssertStruct(t, s.entitySet, s.value)
	}

	m.AssertKey(t, apiProvidersEntitySet, "name", "Edm.String")
	m.AssertKey(t, apiProductsEntitySet, "name", "Edm.String")
	m.AssertKey(t, apiProductAdditionalPropertiesEntity, "entityId", "Edm.String", "name", "Edm.String")
	m.AssertKey(t, certificateStoreReferencesEntitySet, "name", "Edm.String")
	m.AssertKey(t, genericKeyMapEntriesEntitySet, "name", "Edm.String", "scope", "Edm.String", "scopeId", "Edm.String")
}
