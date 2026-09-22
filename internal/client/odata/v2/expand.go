package v2

// ExpandedCollection matches the {"results": [...]} shape SAP's OData V2
// services use for an expanded navigation property nested inside a parent
// entity (via a $expand query option), as opposed to a top-level collection
// response addressed by its own request, which additionally carries
// "__next" for server-driven paging (see DecodeCollectionPage/GetAllPages).
// SAP does not document paging for entries returned through $expand, since
// they are already scoped to a single parent entity rather than an entire
// entity set, so this type intentionally has no next-link field.
type ExpandedCollection[T any] struct {
	Results []T `json:"results"`
}
