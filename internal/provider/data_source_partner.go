package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewPartnerDataSource returns a fresh datasource.DataSource implementation
// for sapintegrationsuite_partner.
func NewPartnerDataSource() datasource.DataSource {
	return &partnerDataSource{}
}

type partnerDataSource struct {
	client *partnerdirectory.Client
}

type partnerDataSourceModel struct {
	Pid               types.String `tfsdk:"pid"`
	RuntimeLocationID types.String `tfsdk:"runtime_location_id"`
}

func (d *partnerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner"
}

func (d *partnerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Confirms whether a given Partner ID (Pid) exists in the Partner Directory. " +
			"There is no sapintegrationsuite_partner resource: SAP documents no confirmed create " +
			"operation for Partners (a Pid comes into existence implicitly the first time a " +
			"StringParameter, BinaryParameter, AlternativePartner, AuthorizedUser, or " +
			"UserCredentialParameter references it), and deleting a Pid is documented as cascading " +
			"to every entity that belongs to it — see docs/guides/partner-directory.md.",
		Attributes: map[string]schema.Attribute{
			"runtime_location_id": runtimeLocationDataSourceAttribute(),
			"pid": schema.StringAttribute{
				Required:    true,
				Description: "The Partner ID to look up.",
			},
		},
	}
}

func (d *partnerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *partnerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config partnerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(d.client, config.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	p, err := client.GetPartner(ctx, config.Pid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory partner", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerDataSourceModel{RuntimeLocationID: config.RuntimeLocationID, Pid: types.StringValue(p.Pid)})...)
}
