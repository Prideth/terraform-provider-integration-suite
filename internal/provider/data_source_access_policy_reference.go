package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
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
		Description: "Reads one artifact reference of an access policy, exactly as SAP stores it. " +
			"Because the values are SAP's raw wire constants, this is also the reliable way to learn " +
			"the artifact_type and operator values for a reference created in the UI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier \"<access_policy_id>/<reference_id>\".",
			},
			"access_policy_id": schema.StringAttribute{
				Required:    true,
				Description: "Numeric ID of the access policy the reference belongs to.",
				Validators:  []validator.String{int64StringValidator{}},
			},
			"reference_id": schema.StringAttribute{
				Required:    true,
				Description: "Numeric ID of the reference itself.",
				Validators:  []validator.String{int64StringValidator{}},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the reference.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the reference, or null when it has none.",
			},
			"artifact_type": schema.StringAttribute{
				Computed:    true,
				Description: "Artifact type constant SAP stores in Type, for example \"INTEGRATION_FLOW\".",
			},
			"attribute": schema.StringAttribute{
				Computed:    true,
				Description: "Artifact attribute the condition is evaluated against (ConditionAttribute).",
			},
			"operator": schema.StringAttribute{
				Computed:    true,
				Description: "Condition type SAP stores in ConditionType, for example \"exactString\".",
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Description: "Exact value or regular expression SAP stores in ConditionValue.",
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

	ref, err := d.client.FindAccessPolicyReference(ctx, config.AccessPolicyID.ValueString(), config.ReferenceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite access policy reference", diagnosticDetail(err))
		return
	}
	if ref == nil {
		resp.Diagnostics.AddError("Access policy reference not found",
			"Access policy "+config.AccessPolicyID.ValueString()+" has no artifact reference with ID "+
				config.ReferenceID.ValueString()+".")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, accessPolicyReferenceDataSourceModel{
		ID:             types.StringValue(config.AccessPolicyID.ValueString() + "/" + ref.ID),
		AccessPolicyID: config.AccessPolicyID,
		ReferenceID:    types.StringValue(ref.ID),
		Name:           types.StringValue(ref.Name),
		Description:    stringOrNull(ref.Description),
		ArtifactType:   types.StringValue(ref.Type),
		Attribute:      types.StringValue(ref.ConditionAttribute),
		Operator:       types.StringValue(ref.ConditionType),
		Value:          types.StringValue(ref.ConditionValue),
	})...)
}
