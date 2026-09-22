package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewPartnerStringParameterDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_partner_string_parameter.
func NewPartnerStringParameterDataSource() datasource.DataSource {
	return &partnerStringParameterDataSource{}
}

type partnerStringParameterDataSource struct {
	client *partnerdirectory.Client
}

func (d *partnerStringParameterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_string_parameter"
}

func (d *partnerStringParameterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Partner Directory string parameter by its (partner_id, " +
			"parameter_id) key, without managing it as a resource.",
		Attributes: map[string]schema.Attribute{
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
				Description: "The string parameter's technical ID.",
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Description: "The parameter's text value.",
			},
		},
	}
}

func (d *partnerStringParameterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *partnerStringParameterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config partnerStringParameterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp, err := d.client.GetStringParameter(ctx, config.PartnerID.ValueString(), config.ParameterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory string parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerStringParameterModel{
		ID:          types.StringValue(sp.Pid + "/" + sp.Id),
		PartnerID:   types.StringValue(sp.Pid),
		ParameterID: types.StringValue(sp.Id),
		Value:       types.StringValue(sp.Value),
	})...)
}
