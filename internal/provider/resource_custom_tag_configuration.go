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
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// customTagConfigurationID is the fixed import/state identifier for this
// tenant-wide singleton resource, matching SAP's own confirmed key
// ('CustomTags') for the underlying entity.
const customTagConfigurationID = "CustomTags"

// NewCustomTagConfigurationResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_custom_tag_configuration.
func NewCustomTagConfigurationResource() resource.Resource {
	return &customTagConfigurationResource{}
}

type customTagConfigurationResource struct {
	client *cloudintegration.Client
}

type customTagConfigurationModel struct {
	ID   types.String     `tfsdk:"id"`
	Tags []customTagModel `tfsdk:"tags"`
}

type customTagModel struct {
	Name            types.String `tfsdk:"name"`
	Mandatory       types.Bool   `tfsdk:"mandatory"`
	PermittedValues []string     `tfsdk:"permitted_values"`
}

func (r *customTagConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_tag_configuration"
}

func (r *customTagConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the tenant-wide Custom Tag Configuration: the set of attributes " +
			"(custom tags) integration package owners are asked, or required, to classify their " +
			"packages with. Backed by the public Integration Content OData V2 API " +
			"(CustomTagConfigurations). This is tenant-level configuration, not an independent " +
			"top-level object the way an integration package or a design-time artifact is — SAP " +
			"documents exactly one configuration per tenant, addressed by the fixed key " +
			"\"CustomTags\", so this resource is a singleton: create it once per tenant (or import " +
			"the tenant's existing configuration) and manage its full desired tag list from there. " +
			"Every write (create or update) sends the complete tag list to SAP's documented " +
			"Overwrite=true endpoint — there is no per-tag add/remove operation, and unspecified " +
			"existing tags are understood to be removed, consistent with what \"Overwrite\" means, " +
			"though SAP's documentation does not use the word \"replace\" explicitly. SAP documents " +
			"no delete or clear operation for this entity at all, so this resource's Delete returns " +
			"an explicit error rather than silently doing nothing or guessing at an unconfirmed " +
			"clearing mechanism — see docs/guides/custom-tag-configurations.md for the full " +
			"reasoning and how to actually retire a tag configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Always \"CustomTags\" — SAP's confirmed fixed identifier for the " +
					"tenant's one custom tag configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.SetNestedAttribute{
				Required: true,
				Description: "The complete desired set of custom tags. Order carries no meaning " +
					"(SAP's documentation never states it does), so this is a set, not a list: " +
					"reordering tags in configuration never plans a change.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The tag's name, for example \"Owner\". Must be unique within this resource.",
						},
						"mandatory": schema.BoolAttribute{
							Required: true,
							Description: "Whether an integration package owner must set this tag. " +
								"SAP's documented examples always set this explicitly (never omit " +
								"it), so this provider requires it explicitly too rather than " +
								"assuming an unconfirmed default.",
						},
						"permitted_values": schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "An optional closed list of values this tag accepts. Order " +
								"carries no meaning, hence a set, which also means an exact " +
								"duplicate value is not representable — SAP's documentation does " +
								"not confirm whether duplicates are meaningful or accepted, so this " +
								"provider does not allow configuring one. Leave unset for a " +
								"free-text tag.",
						},
					},
				},
			},
		},
	}
}

// ValidateConfig rejects two tags with the same name: SAP's documentation
// does not state whether tag names must be unique, but a tag list with two
// conflicting definitions for the same name is inherently ambiguous, and
// this provider would have no principled way to decide which one SAP
// should end up storing.
func (r *customTagConfigurationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config customTagConfigurationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	seen := make(map[string]bool, len(config.Tags))
	for _, tag := range config.Tags {
		if tag.Name.IsUnknown() || tag.Name.IsNull() {
			continue
		}
		name := tag.Name.ValueString()
		if seen[name] {
			resp.Diagnostics.AddAttributeError(
				pathRoot("tags"),
				"Duplicate tag name",
				"Two tags in this configuration share the name \""+name+"\". Tag names must be unique within a custom tag configuration.",
			)
			continue
		}
		seen[name] = true
	}
}

func (r *customTagConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customTagConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customTagConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetCustomTagConfiguration(ctx, tagsFromModel(plan.Tags)); err != nil {
		resp.Diagnostics.AddError("Failed to create the SAP Integration Suite custom tag configuration", diagnosticDetail(err))
		return
	}

	tags, err := r.client.GetCustomTagConfiguration(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read back the SAP Integration Suite custom tag configuration after create", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, customTagConfigurationToModel(tags))...)
}

func (r *customTagConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customTagConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := r.client.GetCustomTagConfiguration(ctx)
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read the SAP Integration Suite custom tag configuration", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, customTagConfigurationToModel(tags))...)
}

func (r *customTagConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customTagConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetCustomTagConfiguration(ctx, tagsFromModel(plan.Tags)); err != nil {
		resp.Diagnostics.AddError("Failed to update the SAP Integration Suite custom tag configuration", diagnosticDetail(err))
		return
	}

	tags, err := r.client.GetCustomTagConfiguration(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read back the SAP Integration Suite custom tag configuration after update", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, customTagConfigurationToModel(tags))...)
}

// Delete deliberately never calls SAP: this project found no documented
// delete or clear operation for CustomTagConfigurations anywhere in SAP's
// public documentation (unlike, for example, integration packages or
// design-time artifacts, which all have a confirmed DELETE). Guessing that
// posting an empty tag list means "delete", or silently removing this
// resource from Terraform state while leaving the tenant's configuration
// untouched, would both misrepresent what actually happened on the
// tenant — so this returns a clear, actionable error instead. See
// docs/guides/custom-tag-configurations.md for how to actually retire a
// tag configuration (SAP's Manage Integration Content UI, or explicitly
// applying an empty tags = [] configuration if your organization has
// separately confirmed that behaves the way you expect against your own
// tenant).
func (r *customTagConfigurationResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Destroying sapintegrationsuite_custom_tag_configuration is not supported",
		"SAP documents no delete or clear operation for the CustomTagConfigurations API, and this "+
			"provider does not guess at one for tenant-wide governance configuration. To stop "+
			"managing this configuration with Terraform without changing anything on the tenant, "+
			"remove it from state with 'terraform state rm' instead of running 'terraform destroy'. "+
			"To actually clear all tags, do so through the SAP Integration Suite UI, or set "+
			"tags = [] in configuration and apply if you have independently confirmed that an "+
			"empty overwrite behaves the way you expect against your own tenant. See "+
			"docs/guides/custom-tag-configurations.md.",
	)
}

func (r *customTagConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != customTagConfigurationID {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"sapintegrationsuite_custom_tag_configuration's import ID must be the literal string "+
				"\"CustomTags\" (SAP's fixed identifier for the tenant's one custom tag configuration), got "+req.ID,
		)
		return
	}
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

func tagsFromModel(tags []customTagModel) []cloudintegration.CustomTag {
	out := make([]cloudintegration.CustomTag, 0, len(tags))
	for _, t := range tags {
		out = append(out, cloudintegration.CustomTag{
			Name:            t.Name.ValueString(),
			Mandatory:       t.Mandatory.ValueBool(),
			PermittedValues: t.PermittedValues,
		})
	}
	return out
}

func customTagConfigurationToModel(tags []cloudintegration.CustomTag) customTagConfigurationModel {
	modelTags := make([]customTagModel, 0, len(tags))
	for _, t := range tags {
		modelTags = append(modelTags, customTagModel{
			Name:            types.StringValue(t.Name),
			Mandatory:       types.BoolValue(t.Mandatory),
			PermittedValues: t.PermittedValues,
		})
	}
	return customTagConfigurationModel{
		ID:   types.StringValue(customTagConfigurationID),
		Tags: modelTags,
	}
}
