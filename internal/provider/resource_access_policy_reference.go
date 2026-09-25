package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ArtifactType   types.String `tfsdk:"artifact_type"`
	Attribute      types.String `tfsdk:"attribute"`
	Operator       types.String `tfsdk:"operator"`
	Value          types.String `tfsdk:"value"`
}

// Values that earlier provider releases accepted for a concept whose real
// wire value is now known. Only provably wrong values are listed; rejecting
// them at plan time replaces an opaque OData error at apply time.
var (
	legacyArtifactTypeValues = map[string]string{
		"IntegrationFlow": `use "INTEGRATION_FLOW"`,
	}
	legacyOperatorValues = map[string]string{
		"EQUALS": `use "exactString"`,
	}
)

func (r *accessPolicyReferenceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_policy_reference"
}

func (r *accessPolicyReferenceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "One artifact reference of an access policy: a rule that says which artifacts " +
			"(by type, and by name or ID) the policy protects. A policy usually has several. Each " +
			"reference is its own ArtifactReferences entity with a server-assigned ID, which is " +
			"why it is a separate resource and not a block inside sapintegrationsuite_access_policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier \"<access_policy_id>/<reference_id>\", both numeric SAP IDs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_policy_id": schema.StringAttribute{
				Required:    true,
				Description: "Numeric ID of the access policy this reference belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					int64StringValidator{},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the reference as shown in the policy's References table. Mandatory in SAP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description, for example what a regular expression is meant to match.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"artifact_type": schema.StringAttribute{
				Required: true,
				Description: "Artifact type constant as SAP's API stores it in the Type property, for " +
					"example \"INTEGRATION_FLOW\". Passed through unchanged; see the Access Policies " +
					"guide for how to find the constant for other types.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					legacyValueValidator{legacy: legacyArtifactTypeValues},
				},
			},
			"attribute": schema.StringAttribute{
				Required: true,
				Description: "Artifact attribute the condition is evaluated against, as stored in " +
					"ConditionAttribute, for example \"Name\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"operator": schema.StringAttribute{
				Required: true,
				Description: "Condition type as stored in ConditionType: \"exactString\" for an exact " +
					"match. The regular-expression variant is covered in the Access Policies guide.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					legacyValueValidator{legacy: legacyOperatorValues},
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "Exact name/ID, or Java regular expression, stored in ConditionValue.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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
	if !requireHTTPClient(data, "resource", &resp.Diagnostics) {
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
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Type:               plan.ArtifactType.ValueString(),
		ConditionAttribute: plan.Attribute.ValueString(),
		ConditionType:      plan.Operator.ValueString(),
		ConditionValue:     plan.Value.ValueString(),
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

	ref, err := r.client.FindAccessPolicyReference(ctx, state.AccessPolicyID.ValueString(), referenceIDFrom(state.ID.ValueString()))
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite access policy reference", diagnosticDetail(err))
		return
	}
	if ref == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, referenceToModel(state.AccessPolicyID.ValueString(), ref))...)
}

// Update is unreachable: every attribute forces replacement, because SAP's
// public tooling only creates and deletes references and no in-place update
// contract for ArtifactReferences has been confirmed.
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

	err := r.client.DeleteAccessPolicyReference(ctx, referenceIDFrom(state.ID.ValueString()))
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
	if err == nil {
		for _, id := range []string{policyID, referenceID} {
			if _, perr := strconv.ParseInt(id, 10, 64); perr != nil {
				err = fmt.Errorf("both parts of %q must be numeric SAP IDs", req.ID)
				break
			}
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			"Expected \"<access_policy_id>/<reference_id>\" with numeric IDs: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("access_policy_id"), policyID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRootID(), policyID+"/"+referenceID)...)
}

func referenceToModel(policyID string, ref *cloudintegration.AccessPolicyReference) accessPolicyReferenceModel {
	return accessPolicyReferenceModel{
		ID:             types.StringValue(policyID + "/" + ref.ID),
		AccessPolicyID: types.StringValue(policyID),
		Name:           types.StringValue(ref.Name),
		Description:    stringOrNull(ref.Description),
		ArtifactType:   types.StringValue(ref.Type),
		Attribute:      types.StringValue(ref.ConditionAttribute),
		Operator:       types.StringValue(ref.ConditionType),
		Value:          types.StringValue(ref.ConditionValue),
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

// int64StringValidator accepts only base-10 integers that fit Edm.Int64.
type int64StringValidator struct{}

func (int64StringValidator) Description(context.Context) string {
	return "value must be a numeric SAP ID"
}

func (v int64StringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (int64StringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := strconv.ParseInt(req.ConfigValue.ValueString(), 10, 64); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid SAP ID",
			fmt.Sprintf("%q is not a numeric SAP ID.", req.ConfigValue.ValueString()))
	}
}

// legacyValueValidator rejects values that earlier provider releases accepted
// by mistake, with a hint where the correct SAP wire value is known.
type legacyValueValidator struct {
	legacy map[string]string
}

func (legacyValueValidator) Description(context.Context) string {
	return "value must be the constant SAP's API uses, not a value accepted by earlier provider releases"
}

func (v legacyValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v legacyValueValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	hint, isLegacy := v.legacy[value]
	if !isLegacy {
		return
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Unsupported legacy value",
		fmt.Sprintf("%q was accepted by earlier releases of this provider, but it is not the value SAP's "+
			"access policy API uses; %s.", value, hint))
}
