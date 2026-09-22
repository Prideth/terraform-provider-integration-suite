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

// NewPartnerUserCredentialParameterResource returns a fresh
// resource.Resource implementation for
// sapintegrationsuite_partner_user_credential_parameter.
func NewPartnerUserCredentialParameterResource() resource.Resource {
	return &partnerUserCredentialParameterResource{}
}

type partnerUserCredentialParameterResource struct {
	client *partnerdirectory.Client
}

// partnerUserCredentialParameterModel carries a PasswordWO field only
// because the framework's reflection-based Plan/State.Get requires every
// schema attribute to have a corresponding struct field to decode into —
// but since password_wo is WriteOnly, Terraform always supplies it as null
// there regardless of what the practitioner configured, and this code
// never reads PasswordWO's value from this struct. The real value is read
// once, directly out of Config (the only place Terraform actually sends
// it), in Create; see the pathRoot("password_wo") lookup there.
// password_wo_version is the plain, stored companion attribute that makes
// password rotation a detectable, plannable change — the standard pattern
// for a write-only secret paired with a version marker.
type partnerUserCredentialParameterModel struct {
	ID                types.String `tfsdk:"id"`
	PartnerID         types.String `tfsdk:"partner_id"`
	ParameterID       types.String `tfsdk:"parameter_id"`
	User              types.String `tfsdk:"user"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.String `tfsdk:"password_wo_version"`
}

func (r *partnerUserCredentialParameterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_user_credential_parameter"
}

func (r *partnerUserCredentialParameterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Partner Directory user credential parameter: a communication " +
			"username and password scoped to a partner ID (Pid), consumed by an integration flow " +
			"through the generated security artifact alias \"pd:<partner_id>:<parameter_id>:" +
			"UserCredential\". Backed by the public Partner Directory OData V2 API " +
			"(UserCredentialParameters). Unlike StringParameters/BinaryParameters, this entity is " +
			"treated as security-sensitive: password_wo is a write-only attribute that Terraform " +
			"never stores in plan or state, this provider never requests or reads a password " +
			"back from SAP, and no in-place update exists — changing the password (via " +
			"password_wo_version), user, partner_id, or parameter_id all replace the resource, " +
			"since no public API for updating an existing credential's password was confirmed. " +
			"UserCredentialParameter cannot be combined with other Partner Directory entity types " +
			"in a single OData batch request; this provider always issues it standalone. See " +
			"docs/guides/partner-directory.md for the full security analysis.",
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
				Description: "The Partner ID (Pid) this credential is scoped to. Does not need to " +
					"already exist: SAP creates it implicitly on the first entity that references it.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parameter_id": schema.StringAttribute{
				Required:    true,
				Description: "The credential parameter's technical ID, unique within its partner.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Required:    true,
				Description: "The communication username.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password_wo": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				WriteOnly: true,
				Description: "The communication password. Write-only: Terraform never stores " +
					"this value in plan or state, and it is never read back from SAP, which does " +
					"not document returning a stored credential's password. Must be supplied on " +
					"every apply that creates or replaces this resource (Terraform requires " +
					"Terraform CLI 1.11 or later for write-only attributes).",
			},
			"password_wo_version": schema.StringAttribute{
				Required: true,
				Description: "An arbitrary value (for example a counter or timestamp) that a " +
					"practitioner changes to signal that password_wo's value has changed and the " +
					"credential should be rotated. Since no in-place password update is confirmed, " +
					"changing this value replaces the resource (delete the old credential, create " +
					"a new one with the new password).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *partnerUserCredentialParameterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *partnerUserCredentialParameterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan partnerUserCredentialParameterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var password types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, pathRoot("password_wo"), &password)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateUserCredentialParameter(ctx,
		plan.PartnerID.ValueString(), plan.ParameterID.ValueString(), plan.User.ValueString(), password.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite Partner Directory user credential parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerUserCredentialParameterModel{
		ID:                types.StringValue(created.Pid + "/" + created.Id),
		PartnerID:         types.StringValue(created.Pid),
		ParameterID:       types.StringValue(created.Id),
		User:              types.StringValue(created.User),
		PasswordWO:        types.StringNull(),
		PasswordWOVersion: plan.PasswordWOVersion,
	})...)
}

func (r *partnerUserCredentialParameterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state partnerUserCredentialParameterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ucp, err := r.client.GetUserCredentialParameter(ctx, state.PartnerID.ValueString(), state.ParameterID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory user credential parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerUserCredentialParameterModel{
		ID:                types.StringValue(ucp.Pid + "/" + ucp.Id),
		PartnerID:         types.StringValue(ucp.Pid),
		ParameterID:       types.StringValue(ucp.Id),
		User:              types.StringValue(ucp.User),
		PasswordWO:        types.StringNull(),
		PasswordWOVersion: state.PasswordWOVersion,
	})...)
}

// Update is unreachable in practice: every attribute besides
// password_wo_version's own value forces replacement, and
// password_wo_version itself is also RequiresReplace, since no confirmed
// public API exists for updating a stored credential's password in place.
func (r *partnerUserCredentialParameterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"sapintegrationsuite_partner_user_credential_parameter does not support in-place updates; "+
			"Terraform should have replaced this resource instead of updating it.",
	)
}

func (r *partnerUserCredentialParameterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerUserCredentialParameterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUserCredentialParameter(ctx, state.PartnerID.ValueString(), state.ParameterID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite Partner Directory user credential parameter", diagnosticDetail(err))
	}
}

// ImportState only recovers partner_id and parameter_id: password_wo can
// never be recovered (SAP does not return it, and it is write-only in this
// provider's own model even if it did), and password_wo_version is a
// practitioner-chosen marker with no server-side equivalent to read back.
// A configuration applied right after import must supply both, which will
// plan as an update to password_wo_version even though nothing server-side
// actually changes — this is an inherent, documented limitation of
// importing a write-only-secret resource, not a bug.
func (r *partnerUserCredentialParameterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	partnerID, parameterID, err := splitCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("partner_id"), partnerID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("parameter_id"), parameterID)...)
}
