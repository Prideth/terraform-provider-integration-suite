package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

// NewPartnerStringParameterResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_partner_string_parameter.
func NewPartnerStringParameterResource() resource.Resource {
	return &partnerStringParameterResource{}
}

type partnerStringParameterResource struct {
	client *partnerdirectory.Client
}

type partnerStringParameterModel struct {
	ID          types.String `tfsdk:"id"`
	PartnerID   types.String `tfsdk:"partner_id"`
	ParameterID types.String `tfsdk:"parameter_id"`
	Value       types.String `tfsdk:"value"`
}

func (r *partnerStringParameterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_string_parameter"
}

func (r *partnerStringParameterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single Partner Directory string parameter, a named text value " +
			"scoped to a partner ID (Pid). Backed by the public Partner Directory OData V2 API " +
			"(StringParameters). Partner Directory data is stored unencrypted: do not put " +
			"passwords, secrets, private keys, tokens, or other sensitive information in a string " +
			"parameter's value — see docs/guides/partner-directory.md.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<partner_id>/<parameter_id>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"partner_id": schema.StringAttribute{
				Required: true,
				Description: "The Partner ID (Pid) this parameter is scoped to. Does not need to " +
					"already exist: SAP creates it implicitly on the first entity that references it.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parameter_id": schema.StringAttribute{
				Required:    true,
				Description: "The string parameter's technical ID, unique within its partner.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "The parameter's text value.",
			},
		},
	}
}

func (r *partnerStringParameterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	if !requireHTTPClient(data, "resource", &resp.Diagnostics) {
		return
	}
	r.client = partnerdirectory.New(data.HTTPClient, data.Host)
}

func (r *partnerStringParameterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan partnerStringParameterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateStringParameter(ctx, partnerdirectory.StringParameter{
		Pid:   plan.PartnerID.ValueString(),
		Id:    plan.ParameterID.ValueString(),
		Value: plan.Value.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite Partner Directory string parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stringParameterToModel(created))...)
}

func (r *partnerStringParameterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state partnerStringParameterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp, err := r.client.GetStringParameter(ctx, state.PartnerID.ValueString(), state.ParameterID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory string parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stringParameterToModel(sp))...)
}

func (r *partnerStringParameterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan partnerStringParameterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateStringParameter(ctx, plan.PartnerID.ValueString(), plan.ParameterID.ValueString(), plan.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update SAP Integration Suite Partner Directory string parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerStringParameterModel{
		ID:          types.StringValue(plan.PartnerID.ValueString() + "/" + plan.ParameterID.ValueString()),
		PartnerID:   plan.PartnerID,
		ParameterID: plan.ParameterID,
		Value:       plan.Value,
	})...)
}

func (r *partnerStringParameterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerStringParameterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStringParameter(ctx, state.PartnerID.ValueString(), state.ParameterID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite Partner Directory string parameter", diagnosticDetail(err))
	}
}

func (r *partnerStringParameterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	partnerID, parameterID, err := splitCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("partner_id"), partnerID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("parameter_id"), parameterID)...)
}

func stringParameterToModel(sp *partnerdirectory.StringParameter) partnerStringParameterModel {
	return partnerStringParameterModel{
		ID:          types.StringValue(sp.Pid + "/" + sp.Id),
		PartnerID:   types.StringValue(sp.Pid),
		ParameterID: types.StringValue(sp.Id),
		Value:       types.StringValue(sp.Value),
	}
}
