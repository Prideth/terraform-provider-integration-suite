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
// Name and Protocol are confirmed against SAP's own documentation, which
// filters on exactly these two properties (`Name eq '...'`,
// `Protocol eq '...'`). EntryPoints/ApiDefinitions are confirmed nested
// navigation properties, expanded via `$expand=EntryPoints,ApiDefinitions`.
type ServiceEndpoint struct {
	Name           string                               `json:"Name"`
	Protocol       string                               `json:"Protocol"`
	EntryPoints    v2.ExpandedCollection[EntryPoint]    `json:"EntryPoints"`
	APIDefinitions v2.ExpandedCollection[APIDefinition] `json:"ApiDefinitions"`
}

// EntryPoint is one runtime URL exposed for a service endpoint. Name and
// Url are documented as required; Type (DEV/TEST/PROD/SANDBOX) is
// documented as optional. The Url property's JSON casing ("Url", not the
// "URL" SAP's prose documentation uses when describing the property's
// *type*) is confirmed from SAP's own open-source Piper library, which
// parses a live ServiceEndpoints response at
// https://github.com/SAP/jenkins-library/blob/master/cmd/integrationArtifactGetServiceEndpoint.go
// — a stronger source than prose alone for exact wire-format casing.
type EntryPoint struct {
	Name string `json:"Name"`
	URL  string `json:"Url"`
	Type string `json:"Type,omitempty"`
}

// APIDefinition is a link to one machine-readable API definition document
// (OpenAPI, EDMX, WSDL, ...) for a service endpoint. URL and Type are both
// documented as required. This project could not independently confirm the
// "Url" JSON casing for this specific nested type the way it could for
// EntryPoint (no example source touching ApiDefinitions was found), but
// applies it by consistency with EntryPoint and SAP's Pascal-case OData
// convention elsewhere in this same entity.
type APIDefinition struct {
	URL  string `json:"Url"`
	Type string `json:"Type"`
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
			if defs[a].Type != defs[b].Type {
				return defs[a].Type < defs[b].Type
			}
			return defs[a].URL < defs[b].URL
		})
	}
}
