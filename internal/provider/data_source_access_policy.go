package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewAccessPolicyDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_access_policy.
func NewAccessPolicyDataSource() datasource.DataSource {
	return &accessPolicyDataSource{}
}

type accessPolicyDataSource struct {
	client *cloudintegration.Client
}

type accessPolicyDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	RoleName             types.String `tfsdk:"role_name"`
	Description          types.String `tfsdk:"description"`
	ReconciliationStatus types.String `tfsdk:"reconciliation_status"`
}

func (d *accessPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_policy"
}

func (d *accessPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing SAP Integration Suite access policy by its SAP-assigned ID. " +
			"Useful for attaching sapintegrationsuite_access_policy_reference resources to a policy " +
			"this provider does not itself manage.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "SAP-assigned technical ID of the access policy.",
			},
			"role_name": schema.StringAttribute{
				Computed:    true,
				Description: "The role this access policy applies to.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The access policy's free-text description.",
			},
			"reconciliation_status": schema.StringAttribute{
				Computed: true,
				Description: "Replication/reconciliation status of the policy against any " +
					"associated runtime, where SAP reports one.",
			},
		},
	}
}

func (d *accessPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *accessPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config accessPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetAccessPolicy(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite access policy", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, accessPolicyDataSourceModel{
		ID:                   types.StringValue(policy.ID),
		RoleName:             types.StringValue(policy.RoleName),
		Description:          stringOrNull(policy.Description),
		ReconciliationStatus: stringOrNull(policy.ReconciliationStatus),
	})...)
}
