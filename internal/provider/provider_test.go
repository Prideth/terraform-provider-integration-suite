package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProvider_Metadata(t *testing.T) {
	p := New("test")()

	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "sapintegrationsuite" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "sapintegrationsuite")
	}
	if resp.Version != "test" {
		t.Errorf("Version = %q, want %q", resp.Version, "test")
	}
}

func TestProvider_Schema_HasHostAndOAuth(t *testing.T) {
	p := New("test")()

	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	if _, ok := resp.Schema.Attributes["host"]; !ok {
		t.Error("expected a \"host\" attribute in the provider schema")
	}
	if _, ok := resp.Schema.Blocks["oauth"]; !ok {
		t.Error("expected an \"oauth\" block in the provider schema")
	}
}

// TestResources_ImplementConfigureAndImport ensures every resource this
// provider registers supports both Configure (needed for lazy client
// construction) and ImportState (import is a first-class feature for every
// resource in this provider, per docs/resource-design.md).
func TestResources_ImplementConfigureAndImport(t *testing.T) {
	p := New("test")().(interface {
		Resources(context.Context) []func() resource.Resource
	})

	for _, factory := range p.Resources(context.Background()) {
		r := factory()

		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "sapintegrationsuite"}, &metaResp)

		if _, ok := r.(resource.ResourceWithConfigure); !ok {
			t.Errorf("%s does not implement resource.ResourceWithConfigure", metaResp.TypeName)
		}
		if _, ok := r.(resource.ResourceWithImportState); !ok {
			t.Errorf("%s does not implement resource.ResourceWithImportState", metaResp.TypeName)
		}

		var schemaResp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			t.Errorf("%s Schema() produced diagnostics: %v", metaResp.TypeName, schemaResp.Diagnostics)
		}
		if len(schemaResp.Schema.Attributes) == 0 {
			t.Errorf("%s has an empty schema", metaResp.TypeName)
		}
	}
}

func TestDataSources_ImplementConfigure(t *testing.T) {
	p := New("test")().(interface {
		DataSources(context.Context) []func() datasource.DataSource
	})

	for _, factory := range p.DataSources(context.Background()) {
		d := factory()

		var metaResp datasource.MetadataResponse
		d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "sapintegrationsuite"}, &metaResp)

		if _, ok := d.(datasource.DataSourceWithConfigure); !ok {
			t.Errorf("%s does not implement datasource.DataSourceWithConfigure", metaResp.TypeName)
		}
	}
}
