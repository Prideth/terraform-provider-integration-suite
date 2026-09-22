package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewIntegrationPackageDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_integration_package.
func NewIntegrationPackageDataSource() datasource.DataSource {
	return &integrationPackageDataSource{}
}

type integrationPackageDataSource struct {
	client *cloudintegration.Client
}

func (d *integrationPackageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_package"
}

func (d *integrationPackageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Cloud Integration package by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The package's technical ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The package's display name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A free-text description of the package.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The package version reported by SAP.",
			},
		},
	}
}

func (d *integrationPackageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	if !requireHTTPClient(data, "data source", &resp.Diagnostics) {
		return
	}
	d.client = cloudintegration.New(data.HTTPClient, data.Host)
}

func (d *integrationPackageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config integrationPackageModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pkg, err := d.client.GetPackage(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite integration package", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, packageToModel(pkg))...)
}
