package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewPartnersDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_partners.
func NewPartnersDataSource() datasource.DataSource {
	return &partnersDataSource{}
}

type partnersDataSource struct {
	client *partnerdirectory.Client
}

type partnersDataSourceModel struct {
	Pids []types.String `tfsdk:"pids"`
}

func (d *partnersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partners"
}

func (d *partnersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every Partner ID (Pid) known to the tenant's Partner Directory. This " +
			"is the primary brownfield discovery mechanism for Partner Directory content in this " +
			"provider, since Partners has no resource of its own — see " +
			"docs/guides/partner-directory.md. Follows SAP's server-driven paging to return the " +
			"complete list, not just its first page.",
		Attributes: map[string]schema.Attribute{
			"pids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Every Partner ID known to the Partner Directory.",
			},
		},
	}
}

func (d *partnersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *partnersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	partners, err := d.client.ListPartners(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list SAP Integration Suite Partner Directory partners", diagnosticDetail(err))
		return
	}

	pids := make([]types.String, 0, len(partners))
	for _, p := range partners {
		pids = append(pids, types.StringValue(p.Pid))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnersDataSourceModel{Pids: pids})...)
}
