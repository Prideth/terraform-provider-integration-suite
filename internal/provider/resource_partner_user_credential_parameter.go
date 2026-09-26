package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
// directly out of Config (the only place Terraform actually sends it) in
// Create and Update.
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
	RuntimeLocationID types.String `tfsdk:"runtime_location_id"`
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
			"(UserCredentialParameters). password_wo is write-only: Terraform never stores it in " +
			"plan or state, and the provider never reads a password back from SAP. Changing user " +
			"or password_wo_version updates the credential in place with a POST, which SAP " +
			"documents as the update method for this entity (PUT is not supported). Because that " +
			"POST overwrites, create refuses to run when the credential already exists; import " +
			"it instead. See docs/guides/partner-directory.md.",
		Attributes: map[string]schema.Attribute{
			"runtime_location_id": runtimeLocationResourceAttribute(),
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
				Description: "The communication username. Changing it updates the credential in place.",
			},
			"password_wo": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				WriteOnly: true,
				Description: "The communication password. Write-only: Terraform never stores " +
					"this value in plan or state, and it is never read back from SAP. It is sent on " +
					"create and on every update, since SAP's update replaces user and password " +
					"together. Requires Terraform CLI 1.11 or later.",
			},
			"password_wo_version": schema.StringAttribute{
				Required: true,
				Description: "An arbitrary value (for example a counter or timestamp) that you " +
					"change whenever password_wo changes. Terraform cannot see the write-only " +
					"password, so changing this value is what triggers the in-place update that " +
					"sends the new password.",
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

func userCredentialParameterToModel(ucp *partnerdirectory.UserCredentialParameter, version, location types.String) partnerUserCredentialParameterModel {
	return partnerUserCredentialParameterModel{
		ID:                types.StringValue(ucp.Pid + "/" + ucp.Id),
		PartnerID:         types.StringValue(ucp.Pid),
		ParameterID:       types.StringValue(ucp.Id),
		User:              types.StringValue(ucp.User),
		PasswordWO:        types.StringNull(),
		PasswordWOVersion: version,
		RuntimeLocationID: location,
	}
}

func (r *partnerUserCredentialParameterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan partnerUserCredentialParameterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, plan.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	var password types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, pathRoot("password_wo"), &password)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pid, id := plan.PartnerID.ValueString(), plan.ParameterID.ValueString()

	// SAP's POST overwrites an existing credential with the same Pid and
	// Id, so creating over one would silently replace a password this
	// configuration does not own.
	if _, err := client.GetUserCredentialParameter(ctx, pid, id); err == nil {
		resp.Diagnostics.AddError(
			"Partner Directory user credential parameter already exists",
			"A user credential parameter "+pid+"/"+id+" already exists. Creating it would overwrite "+
				"its user and password, so the provider stops here. Import it with "+
				"terraform import, then apply to set the configured password.",
		)
		return
	} else if !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to check for an existing SAP Integration Suite Partner Directory user credential parameter", diagnosticDetail(err))
		return
	}

	created, err := client.CreateUserCredentialParameter(ctx, pid, id, plan.User.ValueString(), password.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite Partner Directory user credential parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userCredentialParameterToModel(created, plan.PasswordWOVersion, plan.RuntimeLocationID))...)
}

func (r *partnerUserCredentialParameterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state partnerUserCredentialParameterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, state.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	ucp, err := client.GetUserCredentialParameter(ctx, state.PartnerID.ValueString(), state.ParameterID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory user credential parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userCredentialParameterToModel(ucp, state.PasswordWOVersion, state.RuntimeLocationID))...)
}

// Update changes user and password in place. SAP documents POST with the
// same Pid and Id as the update method for this entity; the request always
// carries both, so the configured password is sent on every update.
func (r *partnerUserCredentialParameterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan partnerUserCredentialParameterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, plan.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	var password types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, pathRoot("password_wo"), &password)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := client.UpdateUserCredentialParameter(ctx,
		plan.PartnerID.ValueString(), plan.ParameterID.ValueString(), plan.User.ValueString(), password.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update SAP Integration Suite Partner Directory user credential parameter", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userCredentialParameterToModel(updated, plan.PasswordWOVersion, plan.RuntimeLocationID))...)
}

func (r *partnerUserCredentialParameterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerUserCredentialParameterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, state.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	err := client.DeleteUserCredentialParameter(ctx, state.PartnerID.ValueString(), state.ParameterID.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite Partner Directory user credential parameter", diagnosticDetail(err))
	}
}

// ImportState only recovers partner_id and parameter_id: password_wo can
// never be recovered (SAP does not return it, and it is write-only in this
// provider's own model even if it did), and password_wo_version is a
// practitioner-chosen marker with no server-side equivalent to read back.
// The first apply after import therefore plans an in-place update that
// sends the configured password.
func (r *partnerUserCredentialParameterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	loc, parts, err := splitLocatedImportID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	partnerID, parameterID := parts[0], parts[1]
	setImportedRuntimeLocation(ctx, loc, resp.State.SetAttribute, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("partner_id"), partnerID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("parameter_id"), parameterID)...)
}
