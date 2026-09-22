package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewMessageMappingDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_message_mapping.
func NewMessageMappingDataSource() datasource.DataSource {
	return &messageMappingDataSource{}
}

type messageMappingDataSource struct {
	client *cloudintegration.Client
}

type messageMappingDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	PackageID types.String `tfsdk:"package_id"`
	MappingID types.String `tfsdk:"mapping_id"`
	Name      types.String `tfsdk:"name"`
	Version   types.String `tfsdk:"version"`
}

func (d *messageMappingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_message_mapping"
}

func (d *messageMappingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing reusable Cloud Integration message mapping artifact's " +
			"metadata by ID. This is the package-level artifact, not an integration flow's " +
			"inline/local message mapping configuration — see docs/resource-design.md.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<package_id>/<mapping_id>\".",
			},
			"package_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the integration package the message mapping belongs to.",
			},
			"mapping_id": schema.StringAttribute{
				Required:    true,
				Description: "The message mapping's technical ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The message mapping's display name.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The design-time version SAP reports for this message mapping.",
			},
		},
	}
}

func (d *messageMappingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *messageMappingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config messageMappingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapping, err := d.client.GetMessageMapping(ctx, config.PackageID.ValueString(), config.MappingID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite message mapping", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, messageMappingDataSourceModel{
		ID:        types.StringValue(config.PackageID.ValueString() + "/" + mapping.ID),
		PackageID: config.PackageID,
		MappingID: types.StringValue(mapping.ID),
		Name:      types.StringValue(mapping.Name),
		Version:   types.StringValue(mapping.Version),
	})...)
}
