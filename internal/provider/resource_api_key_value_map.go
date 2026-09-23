package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apimanagementclassic"
)

// NewAPIKeyValueMapResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_api_key_value_map.
func NewAPIKeyValueMapResource() resource.Resource {
	return &apiKeyValueMapResource{}
}

type apiKeyValueMapResource struct {
	client *apimanagementclassic.Client
}

type apiKeyValueMapModel struct {
	ID        types.String               `tfsdk:"id"`
	Name      types.String               `tfsdk:"name"`
	Scope     types.String               `tfsdk:"scope"`
	ScopeID   types.String               `tfsdk:"scope_id"`
	Entries   []apiKeyValueMapEntryModel `tfsdk:"entries"`
	Encrypted types.Bool                 `tfsdk:"encrypted"`
}

type apiKeyValueMapEntryModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (r *apiKeyValueMapResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key_value_map"
}

func (r *apiKeyValueMapResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Classic API Management Key Value Map (GenericKeyMapEntries): a " +
			"named collection of key/value pairs scoped to an API proxy, an environment, or an " +
			"organization, readable at runtime through the Key Value Map Operations policy. Backed " +
			"by the API Portal's Management.svc OData API, confirmed field-for-field against SAP's " +
			"own documented Create worked example.\n\n" +
			"IMPORTANT — this resource has no Update: SAP's documentation confirms a full Create/" +
			"Update-entries/Delete UI lifecycle exists, but only Create's REST payload is shown " +
			"verbatim anywhere reachable; this provider does not guess at an unconfirmed per-entry " +
			"PUT/POST/DELETE contract. Every attribute, including entries, is RequiresReplace: " +
			"changing anything about a key value map recreates it (delete the old map, create a new " +
			"one with the desired entries) rather than attempt an in-place update this provider " +
			"cannot confirm is safe.\n\n" +
			"IMPORTANT — encrypted maps are not supported. SAP's \"Encrypted\" checkbox (wire field " +
			"isEncrypted) is real and documented, but this provider could not confirm whether GET " +
			"returns an encrypted entry's plaintext value back, a masked placeholder, or nothing at " +
			"all — and getting this wrong would either leak a secret into Terraform state or produce " +
			"permanent diffs. This resource always sends isEncrypted = false and rejects a " +
			"configuration that implies otherwise; store secret values outside this resource until " +
			"that behavior is confirmed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "A composite of name, scope, and scope_id, matching this entity's confirmed OData key.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "The key value map's name.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"scope": schema.StringAttribute{
				Required: true,
				Description: "The map's scope. The only value this provider found confirmed in a " +
					"worked example is \"APIPROXY\"; not validated against a closed enum, since SAP's " +
					"full set of accepted values (for example an environment- or organization-level " +
					"scope) is not confirmed.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"scope_id": schema.StringAttribute{
				Required: true,
				Description: "The identifier of the object scope refers to — for scope = \"APIPROXY\", " +
					"the API proxy's name.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"entries": schema.ListNestedAttribute{
				Optional:    true,
				Description: "The map's key/value entries, sent together with the map itself on Create. See the resource description above for why changing this list replaces the whole resource.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":   schema.StringAttribute{Required: true},
						"value": schema.StringAttribute{Required: true, Sensitive: true},
					},
				},
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"encrypted": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Description: "Must be false (the default). See the resource description above for " +
					"why encrypted maps are not supported.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *apiKeyValueMapResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config struct {
		Encrypted types.Bool `tfsdk:"encrypted"`
	}
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, pathRoot("encrypted"), &config.Encrypted)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.Encrypted.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			pathRoot("encrypted"),
			"Encrypted Key Value Maps are not supported",
			"This provider does not support encrypted Key Value Maps — see the resource description for why. Set encrypted = false, or manage this Key Value Map outside Terraform.",
		)
	}
}

func (r *apiKeyValueMapResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	if !requireAPIManagementClassicHTTPClient(data, "resource", &resp.Diagnostics) {
		return
	}
	r.client = apimanagementclassic.New(data.APIManagementClassicHTTPClient, data.APIManagementClassicHost)
}

func keyValueMapID(name, scope, scopeID string) string {
	return name + "/" + scope + "/" + scopeID
}

func (r *apiKeyValueMapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiKeyValueMapModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kvm := apimanagementclassic.KeyValueMap{
		Name:    plan.Name.ValueString(),
		Scope:   plan.Scope.ValueString(),
		ScopeID: plan.ScopeID.ValueString(),
	}
	for _, e := range plan.Entries {
		kvm.Entries = append(kvm.Entries, apimanagementclassic.KeyValueMapEntry{
			Name: e.Key.ValueString(), Value: e.Value.ValueString(),
		})
	}

	created, err := r.client.CreateKeyValueMap(ctx, kvm)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Classic API Management key value map", diagnosticDetail(err))
		return
	}

	plan.ID = types.StringValue(keyValueMapID(created.Name, created.Scope, created.ScopeID))
	plan.Encrypted = types.BoolValue(created.IsEncrypted)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyValueMapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiKeyValueMapModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.client.GetKeyValueMap(ctx, state.Name.ValueString(), state.Scope.ValueString(), state.ScopeID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Classic API Management key value map", diagnosticDetail(err))
		return
	}

	state.Name = types.StringValue(found.Name)
	state.Scope = types.StringValue(found.Scope)
	state.ScopeID = types.StringValue(found.ScopeID)
	state.ID = types.StringValue(keyValueMapID(found.Name, found.Scope, found.ScopeID))
	state.Encrypted = types.BoolValue(found.IsEncrypted)

	entries := make([]apiKeyValueMapEntryModel, 0, len(found.Entries))
	for _, e := range found.Entries {
		entries = append(entries, apiKeyValueMapEntryModel{Key: types.StringValue(e.Name), Value: types.StringValue(e.Value)})
	}
	state.Entries = entries

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is never called: every attribute in Schema is RequiresReplace, so
// the Terraform Plugin Framework always routes a change through
// Delete+Create instead. This method exists only to satisfy the
// resource.Resource interface.
func (r *apiKeyValueMapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiKeyValueMapModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyValueMapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiKeyValueMapModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteKeyValueMap(ctx, state.Name.ValueString(), state.Scope.ValueString(), state.ScopeID.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete Classic API Management key value map", diagnosticDetail(err))
	}
}

// ImportState accepts "<name>/<scope>/<scope_id>", matching this entity's
// confirmed composite OData key.
func (r *apiKeyValueMapResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected an import ID in the form \"<name>/<scope>/<scope_id>\", got: "+req.ID,
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("scope"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("scope_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("id"), keyValueMapID(parts[0], parts[1], parts[2]))...)
}
