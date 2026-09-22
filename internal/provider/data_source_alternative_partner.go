package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewAlternativePartnerDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_alternative_partner.
func NewAlternativePartnerDataSource() datasource.DataSource {
	return &alternativePartnerDataSource{}
}

type alternativePartnerDataSource struct {
	client *partnerdirectory.Client
}

func (d *alternativePartnerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alternative_partner"
}

func (d *alternativePartnerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Partner Directory alternative partner mapping by its " +
			"(agency, scheme, external_id) tuple, without managing it as a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<hex_agency>/<hex_scheme>/<hex_external_id>\".",
			},
			"agency": schema.StringAttribute{
				Required:    true,
				Description: "The external scheme agency identifier.",
			},
			"scheme": schema.StringAttribute{
				Required:    true,
				Description: "The external identification scheme.",
			},
			"external_id": schema.StringAttribute{
				Required:    true,
				Description: "The external identifier within agency/scheme.",
			},
			"partner_id": schema.StringAttribute{
				Computed:    true,
				Description: "The internal Partner ID (Pid) this external identity resolves to.",
			},
		},
	}
}

func (d *alternativePartnerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alternativePartnerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config alternativePartnerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ap, err := d.client.GetAlternativePartner(ctx, config.Agency.ValueString(), config.Scheme.ValueString(), config.ExternalID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite alternative partner", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, alternativePartnerToModel(ap))...)
}
