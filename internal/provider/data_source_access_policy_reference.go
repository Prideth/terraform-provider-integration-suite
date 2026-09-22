package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewAccessPolicyReferenceDataSource returns a fresh datasource.DataSource
// implementation for sapintegrationsuite_access_policy_reference.
func NewAccessPolicyReferenceDataSource() datasource.DataSource {
	return &accessPolicyReferenceDataSource{}
}

type accessPolicyReferenceDataSource struct {
	client *cloudintegration.Client
}

type accessPolicyReferenceDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	AccessPolicyID types.String `tfsdk:"access_policy_id"`
	ReferenceID    types.String `tfsdk:"reference_id"`
	ArtifactType   types.String `tfsdk:"artifact_type"`
	Attribute      types.String `tfsdk:"attribute"`
	Operator       types.String `tfsdk:"operator"`
	Value          types.String `tfsdk:"value"`
}

func (d *accessPolicyReferenceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_policy_reference"
}

func (d *accessPolicyReferenceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a single existing artifact reference attached to an SAP Integration Suite " +
			"access policy, by the policy's ID and the reference's own SAP-assigned ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<access_policy_id>/<reference_id>\".",
			},
			"access_policy_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the sapintegrationsuite_access_policy this reference belongs to.",
			},
			"reference_id": schema.StringAttribute{
				Required:    true,
				Description: "The reference's own SAP-assigned technical ID.",
			},
			"artifact_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of artifact this reference protects.",
			},
			"attribute": schema.StringAttribute{
				Computed:    true,
				Description: "The artifact attribute this reference matches on.",
			},
			"operator": schema.StringAttribute{
				Computed:    true,
				Description: "The match operator this reference uses.",
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Description: "The value or expression the artifact's attribute must satisfy.",
			},
		},
	}
}

func (d *accessPolicyReferenceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *accessPolicyReferenceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config accessPolicyReferenceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref, err := d.client.GetAccessPolicyReference(ctx, config.AccessPolicyID.ValueString(), config.ReferenceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite access policy reference", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, accessPolicyReferenceDataSourceModel{
		ID:             types.StringValue(config.AccessPolicyID.ValueString() + "/" + ref.ID),
		AccessPolicyID: config.AccessPolicyID,
		ReferenceID:    types.StringValue(ref.ID),
		ArtifactType:   types.StringValue(ref.ArtifactType),
		Attribute:      types.StringValue(ref.Attribute),
		Operator:       types.StringValue(ref.Operator),
		Value:          types.StringValue(ref.Value),
	})...)
}
