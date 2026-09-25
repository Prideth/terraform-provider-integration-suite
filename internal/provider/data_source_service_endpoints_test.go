package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

func serviceEndpointsSchema(t *testing.T) dsschema.Schema {
	t.Helper()
	d := NewServiceEndpointsDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func TestServiceEndpointsDataSource_SchemaOptionalComputed(t *testing.T) {
	s := serviceEndpointsSchema(t)

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"name", false, false},
		{"protocol", false, false},
		{"endpoints", false, true},
	}

	for _, c := range cases {
		attr, ok := s.Attributes[c.name]
		if !ok {
			t.Errorf("missing attribute %q", c.name)
			continue
		}
		if attr.IsRequired() != c.required {
			t.Errorf("%s.Required = %v, want %v", c.name, attr.IsRequired(), c.required)
		}
		if attr.IsComputed() != c.computed {
			t.Errorf("%s.Computed = %v, want %v", c.name, attr.IsComputed(), c.computed)
		}
	}

	if s.Attributes["name"].(dsschema.StringAttribute).Optional != true {
		t.Error("name must be Optional (a filter)")
	}
	if s.Attributes["protocol"].(dsschema.StringAttribute).Optional != true {
		t.Error("protocol must be Optional (a filter)")
	}
}

// The ServiceEndpoint entity type is keyed by Id, so each endpoint exposes
// SAP's own ID as a computed attribute. api_definitions carries name, not the
// never-existing type property earlier releases read.
func TestServiceEndpointsDataSource_NestedAttributesFollowMetadata(t *testing.T) {
	s := serviceEndpointsSchema(t)

	endpoints, ok := s.Attributes["endpoints"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("endpoints is not a ListNestedAttribute")
	}
	for _, name := range []string{"id", "name", "title", "version", "summary", "description", "last_updated", "protocol"} {
		if _, ok := endpoints.NestedObject.Attributes[name]; !ok {
			t.Errorf("endpoints[].%s is missing", name)
		}
	}

	defs, ok := endpoints.NestedObject.Attributes["api_definitions"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("api_definitions is not a ListNestedAttribute")
	}
	if _, ok := defs.NestedObject.Attributes["name"]; !ok {
		t.Error("api_definitions[].name is missing")
	}
	if _, ok := defs.NestedObject.Attributes["type"]; ok {
		t.Error("api_definitions[].type must not exist: the Definition entity has no Type property")
	}
}

func newServiceEndpointsConfig(t *testing.T, s dsschema.Schema, values map[string]tftypes.Value) tfsdk.Config {
	return newFeatureDataSourceConfig(t, s, values)
}

func readServiceEndpoints(t *testing.T, d datasource.DataSourceWithConfigure, client *cloudintegration.Client, values map[string]tftypes.Value) (serviceEndpointsDataSourceModel, datasource.ReadResponse) {
	t.Helper()
	ctx := context.Background()

	// Bypass Configure/requireHTTPClient entirely: these tests point the
	// data source directly at a mock SAP server, not at real provider
	// configuration, so a nil Data.HTTPClient (which Configure would
	// reject) is never relevant here.
	d.(*serviceEndpointsDataSource).client = client

	s := serviceEndpointsSchema(t)
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)},
	}
	d.Read(ctx, datasource.ReadRequest{Config: newServiceEndpointsConfig(t, s, values)}, readResp)

	var state serviceEndpointsDataSourceModel
	readResp.Diagnostics.Append(readResp.State.Get(ctx, &state)...)
	return state, *readResp
}

// TestServiceEndpointsDataSource_Read exercises the full Read path against
// a mock SAP server, checking that nested entry_points/api_definitions are
// mapped correctly, that an optional/absent entry point type comes back
// null rather than an empty string, and that the protocol value is
// preserved exactly as SAP returned it (this provider does not reverse-map
// it).
func TestServiceEndpointsDataSource_Read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": [
			{
				"Id": "ep-customer",
				"Name": "Customer API",
				"Title": "Customer API",
				"LastUpdated": "/Date(1767225600000)/",
				"Protocol": "ODATAV2",
				"EntryPoints": {"results": [{"Name": "default", "Url": "https://tenant.example/http/customer", "Type": "PROD", "AdditionalInformation": "internal only"}]},
				"ApiDefinitions": {"results": [
					{"Url": "https://tenant.example/api/customer.edmx", "Name": "edmx"},
					{"Url": "https://tenant.example/api/customer.json", "Name": "oas-json"}
				]}
			},
			{
				"Name": "Legacy Flow",
				"Protocol": "SOAP",
				"EntryPoints": {"results": [{"Name": "default", "Url": "https://tenant.example/http/legacy"}]},
				"ApiDefinitions": {"results": []}
			}
		]}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)
	d := NewServiceEndpointsDataSource().(datasource.DataSourceWithConfigure)

	state, readResp := readServiceEndpoints(t, d, client, nil)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", readResp.Diagnostics)
	}

	if len(state.Endpoints) != 2 {
		t.Fatalf("len(endpoints) = %d, want 2", len(state.Endpoints))
	}

	customer := state.Endpoints[0]
	if customer.Name.ValueString() != "Customer API" {
		t.Errorf("endpoints[0].name = %q, want \"Customer API\"", customer.Name.ValueString())
	}
	if customer.Protocol.ValueString() != "ODATAV2" {
		t.Errorf("endpoints[0].protocol = %q, want \"ODATAV2\" (preserved exactly, not reverse-mapped)", customer.Protocol.ValueString())
	}
	if len(customer.EntryPoints) != 1 || customer.EntryPoints[0].URL.ValueString() != "https://tenant.example/http/customer" {
		t.Errorf("endpoints[0].entry_points = %+v", customer.EntryPoints)
	}
	if customer.EntryPoints[0].Type.ValueString() != "PROD" {
		t.Errorf("endpoints[0].entry_points[0].type = %q, want PROD", customer.EntryPoints[0].Type.ValueString())
	}
	if len(customer.APIDefinitions) != 2 {
		t.Fatalf("len(endpoints[0].api_definitions) = %d, want 2", len(customer.APIDefinitions))
	}
	if customer.ID.ValueString() != "ep-customer" || customer.Title.ValueString() != "Customer API" {
		t.Errorf("endpoints[0] id/title = %q/%q", customer.ID.ValueString(), customer.Title.ValueString())
	}
	if customer.LastUpdated.ValueString() != "2026-01-01T00:00:00Z" {
		t.Errorf("endpoints[0].last_updated = %q, want RFC 3339 2026-01-01T00:00:00Z", customer.LastUpdated.ValueString())
	}
	if customer.EntryPoints[0].AdditionalInformation.ValueString() != "internal only" {
		t.Errorf("entry_points[0].additional_information = %q", customer.EntryPoints[0].AdditionalInformation.ValueString())
	}
	if customer.APIDefinitions[0].Name.ValueString() != "edmx" {
		t.Errorf("api_definitions[0].name = %q, want edmx", customer.APIDefinitions[0].Name.ValueString())
	}
	if !state.Endpoints[1].LastUpdated.IsNull() {
		t.Errorf("endpoints[1].last_updated = %v, want null when SAP omits it", state.Endpoints[1].LastUpdated)
	}

	legacy := state.Endpoints[1]
	if len(legacy.EntryPoints) != 1 {
		t.Fatalf("len(endpoints[1].entry_points) = %d, want 1", len(legacy.EntryPoints))
	}
	if !legacy.EntryPoints[0].Type.IsNull() {
		t.Errorf("endpoints[1].entry_points[0].type = %q, want null (SAP omitted it)", legacy.EntryPoints[0].Type.ValueString())
	}
	if len(legacy.APIDefinitions) != 0 {
		t.Errorf("len(endpoints[1].api_definitions) = %d, want 0 (empty, not null)", len(legacy.APIDefinitions))
	}
}

// TestServiceEndpointsDataSource_EmptyResult checks that no matching
// endpoints produces a clean, error-free empty list rather than a null
// value or a diagnostic.
func TestServiceEndpointsDataSource_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)
	d := NewServiceEndpointsDataSource().(datasource.DataSourceWithConfigure)

	state, readResp := readServiceEndpoints(t, d, client, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "Does Not Exist"),
	})
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", readResp.Diagnostics)
	}
	if len(state.Endpoints) != 0 {
		t.Errorf("len(endpoints) = %d, want 0", len(state.Endpoints))
	}
}

// TestServiceEndpointsDataSource_NameFilterWithUnicodeAndApostrophe proves
// the name filter reaches the API request correctly encoded for a value
// containing an apostrophe and Unicode characters.
func TestServiceEndpointsDataSource_NameFilterWithUnicodeAndApostrophe(t *testing.T) {
	var gotFilter string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFilter = r.URL.Query().Get("$filter")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"results": []}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)
	d := NewServiceEndpointsDataSource().(datasource.DataSourceWithConfigure)

	name := "Customer's Ünïcödé API"
	_, readResp := readServiceEndpoints(t, d, client, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, name),
	})
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", readResp.Diagnostics)
	}

	want := "Name eq 'Customer''s Ünïcödé API'"
	if gotFilter != want {
		t.Errorf("$filter = %q, want %q", gotFilter, want)
	}
}
