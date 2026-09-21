package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewAccessPolicyReferenceResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_access_policy_reference.
func NewAccessPolicyReferenceResource() resource.Resource {
	return &accessPolicyReferenceResource{}
}

type accessPolicyReferenceResource struct {
	client *cloudintegration.Client
}

type accessPolicyReferenceModel struct {
	ID             types.String `tfsdk:"id"`
	AccessPolicyID types.String `tfsdk:"access_policy_id"`
	ArtifactType   types.String `tfsdk:"artifact_type"`
	Attribute      types.String `tfsdk:"attribute"`
	Operator       types.String `tfsdk:"operator"`
	Value          types.String `tfsdk:"value"`
}

func (r *accessPolicyReferenceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_policy_reference"
}

func (r *accessPolicyReferenceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single artifact reference (artifact type plus a match condition) " +
			"attached to an sapintegrationsuite_access_policy. Modeled as a separate resource " +
			"because each reference has its own server-assigned identity and CRUD lifecycle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<access_policy_id>/<reference_id>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_policy_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the sapintegrationsuite_access_policy this reference belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"artifact_type": schema.StringAttribute{
				Required: true,
				Description: "The type of artifact this reference protects. Must be one of the " +
					"artifact types SAP currently documents as supported.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(cloudintegration.SupportedArtifactTypes...),
				},
			},
			"attribute": schema.StringAttribute{
				Required:    true,
				Description: "The artifact attribute to match on: \"Name\" or \"Id\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("Name", "Id"),
				},
			},
			"operator": schema.StringAttribute{
				Required:    true,
				Description: "The match operator: \"EQUALS\" or \"MATCHES\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("EQUALS", "MATCHES"),
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "The value or expression the artifact's attribute must satisfy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *accessPolicyReferenceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	r.client = cloudintegration.New(data.HTTPClient, data.Host)
}

func (r *accessPolicyReferenceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accessPolicyReferenceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateAccessPolicyReference(ctx, plan.AccessPolicyID.ValueString(), cloudintegration.AccessPolicyReference{
		ArtifactType: plan.ArtifactType.ValueString(),
		Attribute:    plan.Attribute.ValueString(),
		Operator:     plan.Operator.ValueString(),
		Value:        plan.Value.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite access policy reference", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, referenceToModel(plan.AccessPolicyID.ValueString(), created))...)
}

func (r *accessPolicyReferenceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accessPolicyReferenceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref, err := r.client.GetAccessPolicyReference(ctx, state.AccessPolicyID.ValueString(), referenceIDFrom(state.ID.ValueString()))
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite access policy reference", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, referenceToModel(state.AccessPolicyID.ValueString(), ref))...)
}

// Update is unreachable in practice: every attribute besides access_policy_id
// (which forces replacement) is part of the reference's match condition, and
// SAP's artifact reference API does not document in-place updates for it, so
// a changed attribute/operator/value also requires replacement.
func (r *accessPolicyReferenceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"sapintegrationsuite_access_policy_reference does not support in-place updates; "+
			"Terraform should have replaced this resource instead of updating it.",
	)
}

func (r *accessPolicyReferenceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accessPolicyReferenceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAccessPolicyReference(ctx, state.AccessPolicyID.ValueString(), referenceIDFrom(state.ID.ValueString()))
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite access policy reference", diagnosticDetail(err))
	}
}

func (r *accessPolicyReferenceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policyID, referenceID, err := splitCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("access_policy_id"), policyID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRootID(), policyID+"/"+referenceID)...)
}

func referenceToModel(policyID string, ref *cloudintegration.AccessPolicyReference) accessPolicyReferenceModel {
	return accessPolicyReferenceModel{
		ID:             types.StringValue(policyID + "/" + ref.ID),
		AccessPolicyID: types.StringValue(policyID),
		ArtifactType:   types.StringValue(ref.ArtifactType),
		Attribute:      types.StringValue(ref.Attribute),
		Operator:       types.StringValue(ref.Operator),
		Value:          types.StringValue(ref.Value),
	}
}

// referenceIDFrom extracts the reference ID segment from a composite
// "<access_policy_id>/<reference_id>" Terraform ID.
func referenceIDFrom(id string) string {
	_, referenceID, err := splitCompositeID(id)
	if err != nil {
		return id
	}
	return referenceID
}
