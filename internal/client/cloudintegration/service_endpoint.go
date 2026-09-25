package cloudintegration

import (
	"context"
	"sort"
	"strings"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const serviceEndpointsEntitySet = "ServiceEndpoints"

// ServiceEndpoint is the wire representation of a ServiceEndpoints entity
// (Integration Content API): the runtime endpoint(s) SAP exposes for one
// piece of deployed Cloud Integration content. This is a read-only
// discovery resource — SAP generates it from deployed content, and there is
// no create/update/delete operation to call.
//
// The fields follow the ServiceEndpoint entity type of the tenant $metadata:
// Id (key), Name, Title, Version, Summary, Description, LastUpdated
// (Edm.DateTime), Protocol, and the navigation properties EntryPoints and
// ApiDefinitions, expanded via `$expand=EntryPoints,ApiDefinitions`.
type ServiceEndpoint struct {
	ID             string                               `json:"Id"`
	Name           string                               `json:"Name"`
	Title          string                               `json:"Title"`
	Version        string                               `json:"Version"`
	Summary        string                               `json:"Summary"`
	Description    string                               `json:"Description"`
	LastUpdated    string                               `json:"LastUpdated"`
	Protocol       string                               `json:"Protocol"`
	EntryPoints    v2.ExpandedCollection[EntryPoint]    `json:"EntryPoints"`
	APIDefinitions v2.ExpandedCollection[APIDefinition] `json:"ApiDefinitions"`
}

// EntryPoint is one runtime URL exposed for a service endpoint. The tenant
// $metadata defines Name, Url (key), Type (DEV/TEST/PROD/SANDBOX per SAP's
// documentation) and AdditionalInformation.
type EntryPoint struct {
	Name                  string `json:"Name"`
	URL                   string `json:"Url"`
	Type                  string `json:"Type,omitempty"`
	AdditionalInformation string `json:"AdditionalInformation,omitempty"`
}

// APIDefinition is a link to one machine-readable API definition document
// (OpenAPI, EDMX, WSDL, ...) for a service endpoint. The tenant $metadata
// defines this entity type (Definition) with exactly Url (key) and Name.
// SAP's API documentation describes a "Type" with values such as oas-json,
// edmx or wsdl, but no such property exists; earlier releases read it and
// always got an empty value.
type APIDefinition struct {
	URL  string `json:"Url"`
	Name string `json:"Name"`
}

// ListServiceEndpoints lists every service endpoint SAP exposes for
// deployed Cloud Integration content, with EntryPoints and APIDefinitions
// always expanded in the same request (SAP's own example requests
// demonstrate each expansion separately, but OData V2's standard
// comma-separated $expand syntax — already used elsewhere in this
// provider's Query type — combines them in one round trip; no documented
// reason was found that SAP's ServiceEndpoints resource would reject
// that). nameFilter and protocolFilter, when non-empty, apply SAP's
// documented `Name eq '...'`/`Protocol eq '...'` $filter (the only two
// filters SAP documents for this resource), combined with `and` when both
// are supplied. Every page SAP returns is fetched via server-driven
// paging (GetAllPages), and the result is sorted by Name then Protocol —
// and each endpoint's EntryPoints/APIDefinitions sorted too — since SAP
// does not document a guaranteed response order and this provider must
// give Terraform a deterministic list to store in state.
func (c *Client) ListServiceEndpoints(ctx context.Context, nameFilter, protocolFilter string) ([]ServiceEndpoint, error) {
	query := v2.Query{
		Expand: []string{"EntryPoints", "ApiDefinitions"},
		Filter: serviceEndpointsFilter(nameFilter, protocolFilter),
	}
	path := v2.BuildPath(serviceEndpointsEntitySet, "", query.Encode())

	endpoints, err := v2.GetAllPages[ServiceEndpoint](ctx, c.odata, path)
	if err != nil {
		return nil, err
	}

	sortServiceEndpoints(endpoints)
	return endpoints, nil
}

func serviceEndpointsFilter(nameFilter, protocolFilter string) string {
	var parts []string
	if nameFilter != "" {
		parts = append(parts, v2.FilterEquals("Name", nameFilter))
	}
	if protocolFilter != "" {
		parts = append(parts, v2.FilterEquals("Protocol", protocolFilter))
	}
	return strings.Join(parts, " and ")
}

func sortServiceEndpoints(endpoints []ServiceEndpoint) {
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Name != endpoints[j].Name {
			return endpoints[i].Name < endpoints[j].Name
		}
		return endpoints[i].Protocol < endpoints[j].Protocol
	})

	for i := range endpoints {
		entries := endpoints[i].EntryPoints.Results
		sort.Slice(entries, func(a, b int) bool {
			if entries[a].Name != entries[b].Name {
				return entries[a].Name < entries[b].Name
			}
			return entries[a].URL < entries[b].URL
		})

		defs := endpoints[i].APIDefinitions.Results
		sort.Slice(defs, func(a, b int) bool {
			if defs[a].Name != defs[b].Name {
				return defs[a].Name < defs[b].Name
			}
			return defs[a].URL < defs[b].URL
		})
	}
}
