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

// NewPartnerAuthorizedUserResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_partner_authorized_user.
func NewPartnerAuthorizedUserResource() resource.Resource {
	return &partnerAuthorizedUserResource{}
}

type partnerAuthorizedUserResource struct {
	client *partnerdirectory.Client
}

type partnerAuthorizedUserModel struct {
	ID                types.String `tfsdk:"id"`
	User              types.String `tfsdk:"user"`
	PartnerID         types.String `tfsdk:"partner_id"`
	RuntimeLocationID types.String `tfsdk:"runtime_location_id"`
}

func (r *partnerAuthorizedUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_authorized_user"
}

func (r *partnerAuthorizedUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Partner Directory authorized user mapping: which inbound " +
			"communication user is authorized to act as a given Partner ID (Pid). Backed by the " +
			"public Partner Directory OData V2 API (AuthorizedUsers). A communication user maps " +
			"to exactly one Pid, but a Pid can have several authorized users. This resource only " +
			"manages the Partner Directory mapping — it never creates, modifies, or deletes the " +
			"underlying BTP user, OAuth client, or communication user credential itself.",
		Attributes: map[string]schema.Attribute{
			"runtime_location_id": runtimeLocationResourceAttribute(),
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same value as user: the mapping's identity is the communication user itself.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user": schema.StringAttribute{
				Required: true,
				Description: "The communication user this mapping authorizes. Whether SAP " +
					"normalizes this value's case internally has not been confirmed; this provider " +
					"passes it through exactly as configured.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"partner_id": schema.StringAttribute{
				Required: true,
				Description: "The Partner ID (Pid) this user is authorized for. Mutable in place: " +
					"SAP documents PUT for repointing an existing mapping at a different Pid.",
			},
		},
	}
}

func (r *partnerAuthorizedUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *partnerAuthorizedUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan partnerAuthorizedUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, plan.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	created, err := client.CreateAuthorizedUser(ctx, partnerdirectory.AuthorizedUser{
		User: plan.User.ValueString(),
		Pid:  plan.PartnerID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite Partner Directory authorized user", diagnosticDetail(err))
		return
	}

	m := authorizedUserToModel(created)
	m.RuntimeLocationID = plan.RuntimeLocationID
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *partnerAuthorizedUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state partnerAuthorizedUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, state.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	au, err := client.GetAuthorizedUser(ctx, state.User.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite Partner Directory authorized user", diagnosticDetail(err))
		return
	}

	m := authorizedUserToModel(au)
	m.RuntimeLocationID = state.RuntimeLocationID
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *partnerAuthorizedUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan partnerAuthorizedUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, plan.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	err := client.UpdateAuthorizedUser(ctx, plan.User.ValueString(), plan.PartnerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update SAP Integration Suite Partner Directory authorized user", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, partnerAuthorizedUserModel{
		ID:        plan.User,
		User:      plan.User,
		PartnerID: plan.PartnerID,
	})...)
}

func (r *partnerAuthorizedUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerAuthorizedUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, ok := locatedClient(r.client, state.RuntimeLocationID, &resp.Diagnostics)
	if !ok {
		return
	}

	err := client.DeleteAuthorizedUser(ctx, state.User.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite Partner Directory authorized user", diagnosticDetail(err))
	}
}

func (r *partnerAuthorizedUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	loc, parts, err := splitLocatedImportID(req.ID, 1)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("user"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRootID(), parts[0])...)
	setImportedRuntimeLocation(ctx, loc, resp.State.SetAttribute, &resp.Diagnostics)
}

func authorizedUserToModel(au *partnerdirectory.AuthorizedUser) partnerAuthorizedUserModel {
	return partnerAuthorizedUserModel{
		ID:        types.StringValue(au.User),
		User:      types.StringValue(au.User),
		PartnerID: types.StringValue(au.Pid),
	}
}
