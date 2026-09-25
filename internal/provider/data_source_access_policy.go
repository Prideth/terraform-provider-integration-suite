package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	ID          types.String `tfsdk:"id"`
	RoleName    types.String `tfsdk:"role_name"`
	Description types.String `tfsdk:"description"`
}

func (d *accessPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_policy"
}

func (d *accessPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an existing access policy by its numeric ID or by role name. The role " +
			"name lookup is the portable choice: the same policy has a different ID in every tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Numeric ID of the policy. Set exactly one of id or role_name.",
				Validators: []validator.String{
					int64StringValidator{},
				},
			},
			"role_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Role name of the policy, unique per tenant. Set exactly one of id or role_name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The policy's description, or null when it has none.",
			},
		},
	}
}

func (d *accessPolicyDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("role_name")),
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

	var (
		policy *cloudintegration.AccessPolicy
		err    error
	)
	if !config.ID.IsNull() {
		policy, err = d.client.GetAccessPolicy(ctx, config.ID.ValueString())
	} else {
		policy, err = d.client.FindAccessPolicyByRoleName(ctx, config.RoleName.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite access policy", diagnosticDetail(err))
		return
	}
	if policy == nil {
		resp.Diagnostics.AddError("Access policy not found",
			"No access policy with role_name "+config.RoleName.String()+" exists on this tenant.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, accessPolicyDataSourceModel{
		ID:          types.StringValue(policy.ID),
		RoleName:    types.StringValue(policy.RoleName),
		Description: stringOrNull(policy.Description),
	})...)
}
