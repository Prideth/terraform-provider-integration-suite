package partnerdirectory

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
		{alternativePartnersEntitySet, AlternativePartner{}},
		{authorizedUsersEntitySet, AuthorizedUser{}},
		{binaryParametersEntitySet, BinaryParameter{}},
		{partnersEntitySet, Partner{}},
		{stringParametersEntitySet, StringParameter{}},
		{userCredentialParametersEntitySet, UserCredentialParameter{}},
		{userCredentialParametersEntitySet, createUserCredentialParameterRequest{}},
	}
	for _, s := range structs {
		m.AssertStruct(t, s.entitySet, s.value)
	}

	m.AssertKey(t, partnersEntitySet, "Pid", "Edm.String")
	m.AssertKey(t, stringParametersEntitySet, "Pid", "Edm.String", "Id", "Edm.String")
	m.AssertKey(t, binaryParametersEntitySet, "Pid", "Edm.String", "Id", "Edm.String")
	m.AssertKey(t, userCredentialParametersEntitySet, "Pid", "Edm.String", "Id", "Edm.String")
	m.AssertKey(t, alternativePartnersEntitySet, "Hexagency", "Edm.String", "Hexscheme", "Edm.String", "Hexid", "Edm.String")
	m.AssertKey(t, authorizedUsersEntitySet, "User", "Edm.String")
}
