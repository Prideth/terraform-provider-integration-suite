package cloudintegration

import (
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
