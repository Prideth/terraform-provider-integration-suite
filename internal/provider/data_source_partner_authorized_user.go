package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewPartnerAuthorizedUserDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_partner_authorized_user.
func NewPartnerAuthorizedUserDataSource() datasource.DataSource {
	return &partnerAuthorizedUserDataSource{}
}

type partnerAuthorizedUserDataSource struct {
	client *partnerdirectory.Client
}

func (d *partnerAuthorizedUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_authorized_user"
}

func (d *partnerAuthorizedUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Partner Directory authorized user mapping by its " +
			"communication user, without managing it as a resource.",
		Attributes: map[string]schema.Attribute{
			"runtime_location_id": runtimeLocationDataSourceAttribute(),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same value as user.",
			},
			"user": schema.StringAttribute{
				Required:    true,
				Description: "The communication user to look up.",
			},
			"partner_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Partner ID (Pid) this user is authorized for.",
			},
		},
	}
}

func (d *partnerAuthorizedUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *partnerAuthorizedUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config partnerAuthorizedUserModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(d.client, config.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	au, err := client.GetAuthorizedUser(ctx, config.User.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory authorized user", diagnosticDetail(err))
		return
	}

	m := authorizedUserToModel(au)
	m.RuntimeLocationID = config.RuntimeLocationID
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}
