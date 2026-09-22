package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/features"
)

// NewProviderFeatureDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_provider_feature.
func NewProviderFeatureDataSource() datasource.DataSource {
	return &providerFeatureDataSource{}
}

// providerFeatureDataSource, like providerFeaturesDataSource, answers
// entirely from features.Catalog and needs no SAP client.
type providerFeatureDataSource struct {
	version string
}

type providerFeatureDataSourceModel struct {
	ProviderVersion types.String `tfsdk:"provider_version"`
	featureModel
}

func (d *providerFeatureDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_feature"
}

func (d *providerFeatureDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := featureAttributes()
	attrs["key"] = schema.StringAttribute{
		Required: true,
		Description: "The feature key to look up, for example \"cloud_integration.value_mapping\". " +
			"Query sapintegrationsuite_provider_features to list every valid key.",
	}
	attrs["provider_version"] = schema.StringAttribute{
		Computed:    true,
		Description: "The version of this provider that produced this feature's support information.",
	}

	resp.Schema = schema.Schema{
		Description: "Looks up support information for exactly one SAP Integration Suite feature " +
			"by key, from this provider version's built-in feature catalog. Like " +
			"sapintegrationsuite_provider_features, this requires no SAP host or credentials — see " +
			"docs/feature-support.md.",
		Attributes: attrs,
	}
}

func (d *providerFeatureDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	// Deliberately does not call requireHTTPClient — see providerFeaturesDataSource.Configure.
	d.version = data.Version
}

func (d *providerFeatureDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Only "key" is read from the config here, not the whole struct: every
	// other attribute is Computed, so a real Terraform request sends them
	// as null, which a plain (non-pointer, non-types.Object) nested
	// operationsModel struct cannot decode. Decoding just the one
	// practitioner-supplied attribute avoids that entirely.
	var key types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, pathRoot("key"), &key)...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, ok := features.Lookup(key.ValueString())
	if !ok {
		resp.Diagnostics.AddAttributeError(
			pathRoot("key"),
			"Unknown feature key",
			fmt.Sprintf(
				"%q is not a feature key this provider version knows about. "+
					"Use the sapintegrationsuite_provider_features data source to list all %d valid keys, "+
					"or see docs/feature-support.md.",
				key.ValueString(), len(features.Catalog),
			),
		)
		return
	}

	model := providerFeatureDataSourceModel{
		ProviderVersion: types.StringValue(d.version),
		featureModel:    featureToModel(f),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
