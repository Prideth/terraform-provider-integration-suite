package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/features"
)

// NewProviderFeaturesDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_provider_features.
func NewProviderFeaturesDataSource() datasource.DataSource {
	return &providerFeaturesDataSource{}
}

// providerFeaturesDataSource has no client and no Configure-time
// dependency on provider.Data at all: it answers entirely from
// features.Catalog, which is compiled into this binary. See
// docs/feature-support.md for why this must never make a network call.
type providerFeaturesDataSource struct {
	version string
}

type providerFeaturesDataSourceModel struct {
	ID              types.String   `tfsdk:"id"`
	ProviderVersion types.String   `tfsdk:"provider_version"`
	Features        []featureModel `tfsdk:"features"`
}

func (d *providerFeaturesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_features"
}

func (d *providerFeaturesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns this provider version's complete, built-in SAP Integration Suite " +
			"feature support catalog: what this provider implements, what it does not, and why. " +
			"This is static provider metadata baked into the provider binary — it answers \"what " +
			"does this provider version support\", never \"what is activated in my SAP tenant\", " +
			"and requires no SAP host or credentials. See docs/feature-support.md.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always \"provider_features\"; present because Terraform data sources require an id attribute.",
			},
			"provider_version": schema.StringAttribute{
				Computed:    true,
				Description: "The version of this provider that produced this feature catalog.",
			},
			"features": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Every SAP Integration Suite feature this project has evaluated, supported or not.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: featureAttributes(),
				},
			},
		},
	}
}

func (d *providerFeaturesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	// Deliberately does not call requireHTTPClient: this data source
	// answers from features.Catalog alone and must work with no SAP
	// connectivity configured at all.
	d.version = data.Version
}

func (d *providerFeaturesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	models := make([]featureModel, 0, len(features.Catalog))
	for _, f := range features.Catalog {
		models = append(models, featureToModel(f))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, providerFeaturesDataSourceModel{
		ID:              types.StringValue("provider_features"),
		ProviderVersion: types.StringValue(d.version),
		Features:        models,
	})...)
}
