package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewIntegrationAdapterDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_integration_adapter.
func NewIntegrationAdapterDataSource() datasource.DataSource {
	return &integrationAdapterDataSource{}
}

type integrationAdapterDataSource struct {
	client *cloudintegration.Client
}

type integrationAdapterDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PackageID   types.String `tfsdk:"package_id"`
	Description types.String `tfsdk:"description"`
	Version     types.String `tfsdk:"version"`
}

func (d *integrationAdapterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_adapter"
}

func (d *integrationAdapterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing custom Integration Adapter's metadata by its tenant-wide " +
			"ID. Useful for brownfield discovery before importing a " +
			"sapintegrationsuite_integration_adapter resource, or for referencing an adapter this " +
			"provider does not itself manage from a " +
			"sapintegrationsuite_integration_adapter_deployment resource. The result includes " +
			"package_id, which is what an import of the resource needs. See " +
			"docs/guides/integration-adapters.md.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The integration adapter's tenant-wide-unique technical ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The adapter's display name.",
			},
			"package_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the integration package the adapter belongs to.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description SAP took from the *.esa file on import.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The version SAP reports for the adapter.",
			},
		},
	}
}

func (d *integrationAdapterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *integrationAdapterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config integrationAdapterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adapter, err := d.client.GetIntegrationAdapter(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite integration adapter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, integrationAdapterDataSourceModel{
		ID:          types.StringValue(adapter.ID),
		Name:        types.StringValue(adapter.Name),
		PackageID:   stringOrNull(adapter.PackageID),
		Description: stringOrNull(adapter.Description),
		Version:     stringOrNull(adapter.Version),
	})...)
}
