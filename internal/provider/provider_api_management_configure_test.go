package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	provschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// apiManagementProviderConfig builds a provider config with every attribute
// null except the four api_management leaf attributes, mirroring
// emptyProviderConfig's approach of building an explicit-null object value
// per attribute rather than relying on a fully-null root object.
func apiManagementProviderConfig(t *testing.T, s provschema.Schema, host, tokenURL, clientID, clientSecret string) tfsdk.Config {
	t.Helper()
	ctx := context.Background()
	objType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("provider schema type is not a tftypes.Object")
	}

	attrValues := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, at := range objType.AttributeTypes {
		if name != "api_management" {
			attrValues[name] = tftypes.NewValue(at, nil)
			continue
		}
		apimObjType, ok := at.(tftypes.Object)
		if !ok {
			t.Fatalf("api_management attribute type is not a tftypes.Object")
		}
		apimValues := map[string]tftypes.Value{
			"host":          tftypes.NewValue(tftypes.String, stringOrNilValue(host)),
			"token_url":     tftypes.NewValue(tftypes.String, stringOrNilValue(tokenURL)),
			"client_id":     tftypes.NewValue(tftypes.String, stringOrNilValue(clientID)),
			"client_secret": tftypes.NewValue(tftypes.String, stringOrNilValue(clientSecret)),
		}
		attrValues[name] = tftypes.NewValue(apimObjType, apimValues)
	}

	return tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, attrValues)}
}

func stringOrNilValue(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func providerSchemaForAPIManagementTests(t *testing.T) provschema.Schema {
	t.Helper()
	p := New("test")()
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", schemaResp.Diagnostics)
	}
	return schemaResp.Schema
}

func TestProviderConfigure_APIManagement_PartialConfigErrors(t *testing.T) {
	p := New("test")()
	s := providerSchemaForAPIManagementTests(t)

	req := provider.ConfigureRequest{
		Config: apiManagementProviderConfig(t, s, "https://tenant.apiportal.example.com", "https://tenant.authentication.example.com/oauth/token", "", ""),
	}
	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic error for a partially-configured api_management block")
	}
}

func TestProviderConfigure_APIManagement_CompleteConfigBuildsClient(t *testing.T) {
	p := New("test")()
	s := providerSchemaForAPIManagementTests(t)

	req := provider.ConfigureRequest{
		Config: apiManagementProviderConfig(t, s,
			"https://tenant.apiportal.example.com",
			"https://tenant.authentication.example.com/oauth/token",
			"client-id",
			"client-secret",
		),
	}
	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure() produced diagnostics: %v", resp.Diagnostics)
	}

	data, ok := resp.ResourceData.(*Data)
	if !ok {
		t.Fatalf("ResourceData = %T, want *Data", resp.ResourceData)
	}
	if data.APIManagementClassicHTTPClient == nil {
		t.Error("APIManagementClassicHTTPClient is nil after a fully-configured api_management block")
	}
	if data.APIManagementClassicHost != "https://tenant.apiportal.example.com" {
		t.Errorf("APIManagementClassicHost = %q, want the configured host", data.APIManagementClassicHost)
	}
	// Cloud Integration's own client must remain untouched by api_management
	// configuration, proving the two credential sets are independent.
	if data.HTTPClient != nil {
		t.Error("HTTPClient (Cloud Integration) should remain nil when only api_management is configured")
	}
}

func TestProviderConfigure_APIManagement_AbsentBlockLeavesEverythingElseWorking(t *testing.T) {
	p := New("test")()
	s := providerSchemaForAPIManagementTests(t)

	req := provider.ConfigureRequest{Config: emptyProviderConfig(t, s)}
	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure() with no api_management block produced diagnostics: %v", resp.Diagnostics)
	}

	data, ok := resp.ResourceData.(*Data)
	if !ok {
		t.Fatalf("ResourceData = %T, want *Data", resp.ResourceData)
	}
	if data.APIManagementClassicHTTPClient != nil {
		t.Error("APIManagementClassicHTTPClient should be nil when api_management is entirely absent")
	}
}
