package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewPartnerStringParametersDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_partner_string_parameters.
func NewPartnerStringParametersDataSource() datasource.DataSource {
	return &partnerStringParametersDataSource{}
}

type partnerStringParametersDataSource struct {
	client *partnerdirectory.Client
}

type partnerStringParametersDataSourceModel struct {
	PartnerID types.String                      `tfsdk:"partner_id"`
	Values    []partnerStringParameterListEntry `tfsdk:"values"`
}

type partnerStringParameterListEntry struct {
	ParameterID types.String `tfsdk:"parameter_id"`
	Value       types.String `tfsdk:"value"`
}

func (d *partnerStringParametersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_string_parameters"
}

func (d *partnerStringParametersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every string parameter for a given Partner ID (Pid), for brownfield " +
			"discovery before importing individual sapintegrationsuite_partner_string_parameter " +
			"resources. Follows SAP's server-driven paging: String Parameters is documented as an " +
			"entity set that can hold large numbers of entries per partner, so this always returns " +
			"the complete set, not just the first page.",
		Attributes: map[string]schema.Attribute{
			"partner_id": schema.StringAttribute{
				Required:    true,
				Description: "The Partner ID (Pid) to list string parameters for.",
			},
			"values": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Every string parameter belonging to partner_id.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"parameter_id": schema.StringAttribute{
							Computed:    true,
							Description: "The string parameter's technical ID.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The parameter's text value.",
						},
					},
				},
			},
		},
	}
}

func (d *partnerStringParametersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *partnerStringParametersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config partnerStringParametersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params, err := d.client.ListStringParameters(ctx, config.PartnerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list SAP Integration Suite Partner Directory string parameters", diagnosticDetail(err))
		return
	}

	values := make([]partnerStringParameterListEntry, 0, len(params))
	for _, p := range params {
		values = append(values, partnerStringParameterListEntry{
			ParameterID: types.StringValue(p.Id),
			Value:       types.StringValue(p.Value),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerStringParametersDataSourceModel{
		PartnerID: config.PartnerID,
		Values:    values,
	})...)
}
