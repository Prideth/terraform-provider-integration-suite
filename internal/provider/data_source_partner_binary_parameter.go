package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewPartnerBinaryParameterDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_partner_binary_parameter.
func NewPartnerBinaryParameterDataSource() datasource.DataSource {
	return &partnerBinaryParameterDataSource{}
}

type partnerBinaryParameterDataSource struct {
	client *partnerdirectory.Client
}

type partnerBinaryParameterDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	PartnerID         types.String `tfsdk:"partner_id"`
	ParameterID       types.String `tfsdk:"parameter_id"`
	ContentType       types.String `tfsdk:"content_type"`
	Value             types.String `tfsdk:"value"`
	RuntimeLocationID types.String `tfsdk:"runtime_location_id"`
}

func (d *partnerBinaryParameterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_binary_parameter"
}

func (d *partnerBinaryParameterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Partner Directory binary parameter by its (partner_id, " +
			"parameter_id) key, without managing it as a resource. value is the raw base64-encoded " +
			"content exactly as SAP returns it.",
		Attributes: map[string]schema.Attribute{
			"runtime_location_id": runtimeLocationDataSourceAttribute(),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<partner_id>/<parameter_id>\".",
			},
			"partner_id": schema.StringAttribute{
				Required:    true,
				Description: "The Partner ID (Pid) this parameter is scoped to.",
			},
			"parameter_id": schema.StringAttribute{
				Required:    true,
				Description: "The binary parameter's technical ID.",
			},
			"content_type": schema.StringAttribute{
				Computed:    true,
				Description: "The content's documented type (for example xml, xsd, zip, crt).",
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Description: "The base64-encoded content, exactly as SAP returns it.",
			},
		},
	}
}

func (d *partnerBinaryParameterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = partnerdirectory.New(data.HTTPClient, data.Host)
}

func (d *partnerBinaryParameterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config partnerBinaryParameterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(d.client, config.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	bp, err := client.GetBinaryParameter(ctx, config.PartnerID.ValueString(), config.ParameterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory binary parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerBinaryParameterDataSourceModel{RuntimeLocationID: config.RuntimeLocationID,
		ID:          types.StringValue(bp.Pid + "/" + bp.Id),
		PartnerID:   types.StringValue(bp.Pid),
		ParameterID: types.StringValue(bp.Id),
		ContentType: types.StringValue(bp.ContentType),
		Value:       types.StringValue(bp.Value),
	})...)
}
