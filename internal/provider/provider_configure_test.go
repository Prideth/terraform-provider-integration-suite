package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	provschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// emptyProviderConfig builds the fully-null tfsdk.Config a bare
//
//	provider "sapintegrationsuite" {}
//
// block produces: every attribute (host, oauth.*) unset. This is the case
// this phase's architecture requires to keep working, since the feature
// catalog data sources must be usable with no SAP tenant at all.
func emptyProviderConfig(t *testing.T, s provschema.Schema) tfsdk.Config {
	t.Helper()
	ctx := context.Background()
	objType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("provider schema type is not a tftypes.Object")
	}

	// A null root object (tftypes.NewValue(objType, nil)) is not the same
	// thing as an object whose attributes are all null, and the framework's
	// reflection needs the latter to decode into providerModel. Build an
	// object value with an explicit null for every attribute instead.
	attrValues := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, at := range objType.AttributeTypes {
		attrValues[name] = tftypes.NewValue(at, nil)
	}

	return tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, attrValues)}
}

// TestProviderConfigure_EmptyConfigSucceedsAndYieldsNoHTTPClient is the
// "offline metadata case" this phase's architecture requires: a bare
// provider block with no host or OAuth credentials must configure
// successfully (no diagnostics), producing a *Data with a nil HTTPClient
// rather than failing Configure altogether. Individual SAP-backed
// resources and data sources are responsible for erroring on that nil
// client themselves — see TestProviderConfigure_SAPResourceErrorsWithoutCredentials.
func TestProviderConfigure_EmptyConfigSucceedsAndYieldsNoHTTPClient(t *testing.T) {
	p := New("test")()

	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", schemaResp.Diagnostics)
	}

	req := provider.ConfigureRequest{Config: emptyProviderConfig(t, schemaResp.Schema)}
	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure() with an empty provider block produced diagnostics: %v", resp.Diagnostics)
	}

	data, ok := resp.ResourceData.(*Data)
	if !ok {
		t.Fatalf("ResourceData = %T, want *Data", resp.ResourceData)
	}
	if data.HTTPClient != nil {
		t.Error("HTTPClient is non-nil after Configure() with no SAP host/OAuth credentials")
	}
	if resp.DataSourceData == nil {
		t.Error("DataSourceData was not set")
	}
}

// TestProviderConfigure_SAPResourceErrorsWithoutCredentials is the "SAP
// resource case": a SAP-backed resource configured with the same nil-client
// Data an empty provider block produces must fail Configure with a clear,
// specific diagnostic, rather than silently building a client that could
// never authenticate.
func TestProviderConfigure_SAPResourceErrorsWithoutCredentials(t *testing.T) {
	data := &Data{Version: "test"} // HTTPClient deliberately nil, as if the provider block were empty

	r := NewIntegrationPackageResource().(resource.ResourceWithConfigure)
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: data}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a configuration error for a SAP-backed resource with no SAP credentials")
	}
}

// TestProviderConfigure_FeatureCatalogDataSourcesNeverRequireCredentials
// asserts the flip side directly against the provider's own Configure
// output (rather than a hand-built Data, as in feature_datasource_test.go)
// for both new data sources.
func TestProviderConfigure_FeatureCatalogDataSourcesNeverRequireCredentials(t *testing.T) {
	p := New("test")()

	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)

	var configureResp provider.ConfigureResponse
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: emptyProviderConfig(t, schemaResp.Schema),
	}, &configureResp)
	if configureResp.Diagnostics.HasError() {
		t.Fatalf("provider Configure() produced diagnostics: %v", configureResp.Diagnostics)
	}

	for _, factory := range []func() datasource.DataSource{NewProviderFeaturesDataSource, NewProviderFeatureDataSource} {
		d := factory().(datasource.DataSourceWithConfigure)

		var metaResp datasource.MetadataResponse
		d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "sapintegrationsuite"}, &metaResp)

		var dsConfigureResp datasource.ConfigureResponse
		d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: configureResp.DataSourceData}, &dsConfigureResp)
		if dsConfigureResp.Diagnostics.HasError() {
			t.Errorf("%s Configure() produced diagnostics with no SAP credentials: %v", metaResp.TypeName, dsConfigureResp.Diagnostics)
		}
	}
}
