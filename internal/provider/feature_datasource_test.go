package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/features"
)

// newFeatureDataSourceConfig builds a tfsdk.Config for a data source schema
// with every attribute null except the ones named in values. This is what
// Terraform itself hands a data source's Read method; a plain zero-value
// tfsdk.Config will not do, since Get/GetAttribute need Raw to be a fully
// typed object matching the schema, not an empty tftypes.Value.
func newFeatureDataSourceConfig(t *testing.T, s dsschema.Schema, values map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	ctx := context.Background()

	objType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("schema type is not a tftypes.Object")
	}

	attrValues := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, at := range objType.AttributeTypes {
		if v, set := values[name]; set {
			attrValues[name] = v
			continue
		}
		attrValues[name] = tftypes.NewValue(at, nil)
	}

	return tfsdk.Config{
		Schema: s,
		Raw:    tftypes.NewValue(objType, attrValues),
	}
}

// TestProviderFeaturesDataSource_WorksWithoutSAPCredentials is the core
// architectural guarantee this phase introduces: the feature catalog must
// be queryable with no SAP host or OAuth configuration at all, since it
// describes this provider binary, not any tenant. Configure is called with
// a Data value that has no HTTPClient (exactly what the provider now
// produces when host/OAuth are unset — see provider.go), and Read must
// still succeed.
func TestProviderFeaturesDataSource_WorksWithoutSAPCredentials(t *testing.T) {
	d := NewProviderFeaturesDataSource()

	var schemaResp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", schemaResp.Diagnostics)
	}

	configurable, ok := d.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Fatal("data source does not implement DataSourceWithConfigure")
	}

	var configureResp datasource.ConfigureResponse
	configurable.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: &Data{Version: "test"}, // HTTPClient deliberately nil
	}, &configureResp)
	if configureResp.Diagnostics.HasError() {
		t.Fatalf("Configure() with no SAP credentials produced diagnostics: %v", configureResp.Diagnostics)
	}

	var freshSchemaResp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &freshSchemaResp)

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Schema: freshSchemaResp.Schema, Raw: tftypes.NewValue(freshSchemaResp.Schema.Type().TerraformType(context.Background()), nil)},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: newFeatureDataSourceConfig(t, freshSchemaResp.Schema, nil),
	}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() with no SAP credentials produced diagnostics: %v", readResp.Diagnostics)
	}

	var state providerFeaturesDataSourceModel
	readResp.Diagnostics.Append(readResp.State.Get(context.Background(), &state)...)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("reading back state produced diagnostics: %v", readResp.Diagnostics)
	}

	if state.ProviderVersion.ValueString() != "test" {
		t.Errorf("provider_version = %q, want %q", state.ProviderVersion.ValueString(), "test")
	}
	if len(state.Features) != len(features.Catalog) {
		t.Errorf("got %d features, want %d (len(features.Catalog))", len(state.Features), len(features.Catalog))
	}
}

// TestProviderFeatureDataSource_WorksWithoutSAPCredentials mirrors the
// plural test above for the single-feature lookup data source.
func TestProviderFeatureDataSource_WorksWithoutSAPCredentials(t *testing.T) {
	d := NewProviderFeatureDataSource()

	configurable := d.(datasource.DataSourceWithConfigure)
	var configureResp datasource.ConfigureResponse
	configurable.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: &Data{Version: "test"},
	}, &configureResp)
	if configureResp.Diagnostics.HasError() {
		t.Fatalf("Configure() with no SAP credentials produced diagnostics: %v", configureResp.Diagnostics)
	}

	var schemaResp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", schemaResp.Diagnostics)
	}

	config := newFeatureDataSourceConfig(t, schemaResp.Schema, map[string]tftypes.Value{
		"key": tftypes.NewValue(tftypes.String, "cloud_integration.value_mapping"),
	})

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil)},
	}
	d.Read(context.Background(), datasource.ReadRequest{Config: config}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", readResp.Diagnostics)
	}

	var state providerFeatureDataSourceModel
	readResp.Diagnostics.Append(readResp.State.Get(context.Background(), &state)...)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("reading back state produced diagnostics: %v", readResp.Diagnostics)
	}

	if state.SupportStatus.ValueString() != string(features.StatusPartial) {
		t.Errorf("support_status = %q, want %q", state.SupportStatus.ValueString(), features.StatusPartial)
	}
	if state.Operations.Update.ValueBool() {
		t.Error("operations.update = true for cloud_integration.value_mapping, want false")
	}
}

func TestProviderFeatureDataSource_UnknownKeyIsError(t *testing.T) {
	d := NewProviderFeatureDataSource()

	var schemaResp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)

	config := newFeatureDataSourceConfig(t, schemaResp.Schema, map[string]tftypes.Value{
		"key": tftypes.NewValue(tftypes.String, "does.not.exist"),
	})

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), nil)},
	}
	d.Read(context.Background(), datasource.ReadRequest{Config: config}, readResp)
	if !readResp.Diagnostics.HasError() {
		t.Fatal("expected an error for an unknown feature key")
	}
}
