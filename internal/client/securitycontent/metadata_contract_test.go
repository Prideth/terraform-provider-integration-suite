package securitycontent

import (
	"testing"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/testutil/edmx"
)

// TestWireContractAgainstMetadata checks every wire struct and entity set
// this package uses against a tenant $metadata document. It skips when no
// document is available (see internal/testutil/edmx).
func TestWireContractAgainstMetadata(t *testing.T) {
	m := edmx.Load(t)

	structs := []struct {
		entitySet string
		value     any
	}{
		{oauth2ClientCredentialsEntitySet, OAuth2ClientCredential{}},
		{oauth2ClientCredentialsEntitySet, oauth2ClientCredentialWriteRequest{}},
		{userCredentialsEntitySet, UserCredential{}},
		{userCredentialsEntitySet, userCredentialWriteRequest{}},
		{keystoreEntriesEntitySet, KeystoreEntry{}},
		{keyPairGenerationRequestsEntitySet, keyPairGenerationWireRequest{}},
		{keystoreResourcesEntitySet, deleteKeystoreEntriesRequest{}},
	}
	for _, s := range structs {
		m.AssertStruct(t, s.entitySet, s.value)
	}
	m.AssertKey(t, keystoreEntriesEntitySet, "Hexalias", "Edm.String")
	m.AssertKey(t, certificateResourcesEntitySet, "Hexalias", "Edm.String")
	m.AssertKey(t, oauth2ClientCredentialsEntitySet, "Name", "Edm.String")
	m.AssertKey(t, userCredentialsEntitySet, "Name", "Edm.String")
	m.AssertKey(t, keystoreResourcesEntitySet, "Name", "Edm.String")
}
