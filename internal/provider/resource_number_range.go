package provider

import (
	"context"
	"fmt"
	"math/big"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// numberRangeDigits matches an unsigned decimal integer with no sign, no
// leading/trailing whitespace, and no leading zeros other than "0" itself —
// SAP's wire format for every numeric field on this entity (CurrentValue,
// MinValue, MaxValue, FieldLength) is a JSON string, confirmed verbatim from
// SAP's own documented example bodies.
var numberRangeDigits = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

const numberRangeInvalidValueSummary = "Invalid Number Range value"
const numberRangeGotSuffix = ", got: "

// NewNumberRangeResource returns a fresh resource.Resource implementation
// for sapintegrationsuite_number_range.
func NewNumberRangeResource() resource.Resource {
	return &numberRangeResource{}
}

type numberRangeResource struct {
	client *cloudintegration.Client
}

type numberRangeModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	MinValue              types.String `tfsdk:"min_value"`
	MaxValue              types.String `tfsdk:"max_value"`
	Description           types.String `tfsdk:"description"`
	Rotate                types.Bool   `tfsdk:"rotate"`
	FieldLength           types.String `tfsdk:"field_length"`
	CurrentValueWO        types.String `tfsdk:"current_value_wo"`
	CurrentValueWOVersion types.String `tfsdk:"current_value_wo_version"`
}

func (r *numberRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_number_range"
}

func (r *numberRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the static configuration of a Cloud Integration \"Number Ranges\" " +
			"object, used to generate unique interchange numbers for outbound EDI/EDIFACT " +
			"documents. Backed by the public Message Stores OData V2 API (NumberRanges). " +
			"\n\n" +
			"IMPORTANT — this resource has no Read: SAP documents no GET operation for this " +
			"entity anywhere (unlike DataStores, DataStoreEntries, and Variables, which all have " +
			"documented GET examples in the same API family). Without a GET, this provider cannot " +
			"verify what SAP actually stored, cannot detect drift, and cannot support " +
			"'terraform import'. This resource's Read is a documented no-op that simply trusts " +
			"whatever Terraform already has in state — it never contacts SAP. Every Create and " +
			"Update instead resends the complete static configuration from your configuration " +
			"file, and this provider trusts that write to have succeeded exactly as sent. " +
			"See docs/guides/runtime-stores-and-number-ranges.md for the full reasoning, " +
			"including why 'terraform destroy' is also unsupported (SAP documents no DELETE for " +
			"this entity either — only an 'Undeploy' UI action with no confirmed REST " +
			"equivalent).\n\n" +
			"The runtime counter (SAP's CurrentValue, shown as 'Next Value' in the UI) is " +
			"handled separately from the rest of this resource's configuration: see " +
			"current_value_wo below. Ordinary applies that only change description, min_value, " +
			"max_value, rotate, or field_length never touch the counter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Always equal to name — SAP's confirmed OData key for this entity " +
					"(NumberRanges('<name>')).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "The Number Range object's name. This is SAP's documented OData " +
					"key, so it is RequiresReplace: SAP's documentation does not describe any " +
					"rename semantics, and emulating a rename via Delete+Create is not possible " +
					"anyway since this entity has no confirmed Delete.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"min_value": schema.StringAttribute{
				Required: true,
				Description: "The lowest value the counter may hold, as a decimal digit string " +
					"(SAP's wire format — not a Terraform number, to avoid any numeric-precision " +
					"assumption on values SAP documents as up to 14 digits long). SAP's UI " +
					"validates this as greater than or equal to 0.",
				Validators: []validator.String{numberRangeDigitsValidator{}},
			},
			"max_value": schema.StringAttribute{
				Required: true,
				Description: "The highest value the counter may hold before it errors (if " +
					"rotate is false) or wraps back to min_value (if rotate is true), as a " +
					"decimal digit string. SAP's UI validates this as fewer than 15 digits.",
				Validators: []validator.String{numberRangeDigitsValidator{}, numberRangeMaxDigits15{}},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "A free-text description of the Number Range object. Always sent " +
					"on Create and Update (as an empty string if unset), matching SAP's " +
					"documented example, which always includes this field.",
				Default: stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rotate": schema.BoolAttribute{
				Required: true,
				Description: "Whether the counter wraps back to min_value once it reaches " +
					"max_value (confirmed by SAP's documentation), instead of erroring once " +
					"exhausted. A Terraform bool; SAP's wire format is the string \"true\"/" +
					"\"false\", translated by this provider.",
			},
			"field_length": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The zero-padded display width of the counter value, as a decimal " +
					"digit string (SAP's UI documents a maximum of 14; 0 means no padding is " +
					"applied). Always sent on Create and Update (as \"0\" if unset), matching " +
					"SAP's documented example, which always includes this field.",
				Default:    stringdefault.StaticString("0"),
				Validators: []validator.String{numberRangeFieldLengthValidator{}},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"current_value_wo": schema.StringAttribute{
				Required:  true,
				WriteOnly: true,
				Description: "The counter value (SAP's CurrentValue / the UI's \"Next Value\") " +
					"to push to SAP, as a decimal digit string. Write-only: Terraform never " +
					"stores this in plan or state, and — because this entity has no GET — this " +
					"provider can never read the live counter back either. Required on Create " +
					"(SAP's documented example always includes CurrentValue when adding an " +
					"object). On Update, this value is sent ONLY when current_value_wo_version " +
					"changes from the prior apply; every other Update omits CurrentValue from " +
					"the request entirely, so that changing only description/min_value/" +
					"max_value/rotate/field_length can never reset a counter that has since " +
					"advanced through normal EDI/EDIFACT processing. Setting this on an apply " +
					"that does not also bump current_value_wo_version has no effect — bump the " +
					"version to signal deliberate intent to (re)set the counter, the same " +
					"pattern this provider uses for credential rotation.",
			},
			"current_value_wo_version": schema.StringAttribute{
				Required: true,
				Description: "An arbitrary marker (for example a counter or timestamp) that a " +
					"practitioner changes to signal that current_value_wo should be pushed to " +
					"SAP on this apply. Required on Create (its value is not otherwise used, " +
					"but Create always pushes current_value_wo regardless). Leaving this " +
					"unchanged across an Update is what keeps this resource from ever resetting " +
					"a runtime counter it cannot read back.",
			},
		},
	}
}

func (r *numberRangeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config numberRangeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.MinValue.IsUnknown() || config.MinValue.IsNull() || config.MaxValue.IsUnknown() || config.MaxValue.IsNull() {
		return
	}

	minVal, minOK := new(big.Int).SetString(config.MinValue.ValueString(), 10)
	maxVal, maxOK := new(big.Int).SetString(config.MaxValue.ValueString(), 10)
	if !minOK || !maxOK {
		// Malformed digit strings are already reported by the per-attribute
		// validators; avoid a redundant/confusing second diagnostic.
		return
	}

	if minVal.Cmp(maxVal) > 0 {
		resp.Diagnostics.AddAttributeError(
			pathRoot("min_value"),
			"min_value must not be greater than max_value",
			fmt.Sprintf("min_value (%s) is greater than max_value (%s).", minVal, maxVal),
		)
	}
}

func (r *numberRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *numberRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan numberRangeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var currentValue types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, pathRoot("current_value_wo"), &currentValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := currentValue.ValueString()
	err := r.client.CreateNumberRange(ctx, cloudintegration.NumberRange{
		Name:         plan.Name.ValueString(),
		MinValue:     plan.MinValue.ValueString(),
		MaxValue:     plan.MaxValue.ValueString(),
		Description:  plan.Description.ValueString(),
		Rotate:       plan.Rotate.ValueBool(),
		FieldLength:  plan.FieldLength.ValueString(),
		CurrentValue: &value,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create SAP Integration Suite number range", diagnosticDetail(err))
		return
	}

	plan.ID = types.StringValue(plan.Name.ValueString())
	plan.CurrentValueWO = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read never contacts SAP: no GET operation is documented for this entity
// (see the schema description above and docs/guides/runtime-stores-and-
// number-ranges.md). It simply re-persists whatever Terraform already has
// in state, trusting the last Create/Update to have applied exactly as
// sent — the only contract this provider can honestly offer against this
// API.
func (r *numberRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state numberRangeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *numberRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state numberRangeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nr := cloudintegration.NumberRange{
		Name:        plan.Name.ValueString(),
		MinValue:    plan.MinValue.ValueString(),
		MaxValue:    plan.MaxValue.ValueString(),
		Description: plan.Description.ValueString(),
		Rotate:      plan.Rotate.ValueBool(),
		FieldLength: plan.FieldLength.ValueString(),
	}

	// current_value_wo is only ever pushed when current_value_wo_version
	// changed since the last apply — a deliberate, explicit signal from the
	// practitioner. Every other Update omits CurrentValue entirely, so that
	// changing only static fields can never reset a counter this provider
	// has no way to read back first. See docs/guides/runtime-stores-and-
	// number-ranges.md for why this is the safest available contract given
	// SAP documents no GET for this entity.
	if !plan.CurrentValueWOVersion.Equal(state.CurrentValueWOVersion) {
		var currentValue types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, pathRoot("current_value_wo"), &currentValue)...)
		if resp.Diagnostics.HasError() {
			return
		}
		value := currentValue.ValueString()
		nr.CurrentValue = &value
	}

	if err := r.client.UpdateNumberRange(ctx, nr); err != nil {
		resp.Diagnostics.AddError("Failed to update SAP Integration Suite number range", diagnosticDetail(err))
		return
	}

	plan.ID = types.StringValue(plan.Name.ValueString())
	plan.CurrentValueWO = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deliberately never calls SAP: no DELETE operation is documented
// anywhere for NumberRanges (unlike every other design-time artifact this
// provider manages). SAP's Monitor UI shows an "Undeploy" action for this
// entity, but no REST equivalent was found in SAP's own documentation, and
// this provider does not invent one. Returning an explicit error here
// mirrors sapintegrationsuite_custom_tag_configuration's Delete, the
// established pattern in this codebase for "SAP documents no confirmed
// destroy operation" — see docs/guides/runtime-stores-and-number-ranges.md
// for how to actually retire a Number Range.
func (r *numberRangeResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Destroying sapintegrationsuite_number_range is not supported",
		"SAP documents no delete operation for the NumberRanges API — only an 'Undeploy' action "+
			"in the Monitor UI with no confirmed REST equivalent — and this provider does not "+
			"guess at one for an object that may still be referenced by deployed EDI/EDIFACT "+
			"content. To stop managing this Number Range with Terraform without changing "+
			"anything on the tenant, remove it from state with 'terraform state rm' instead of "+
			"running 'terraform destroy'. To actually remove it, use the SAP Integration Suite "+
			"Monitor UI. See docs/guides/runtime-stores-and-number-ranges.md.",
	)
}

// ImportState always errors: SAP documents no GET operation for this
// entity, so there is nothing this provider could read from the tenant to
// populate min_value, max_value, description, rotate, or field_length with.
// Silently accepting an import ID and leaving every other attribute unknown
// would misrepresent what this provider actually knows, so — mirroring
// Delete above — this returns an explicit, actionable error instead of a
// silently broken import.
func (r *numberRangeResource) ImportState(_ context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError(
		"Importing sapintegrationsuite_number_range is not supported",
		"SAP documents no GET operation for the NumberRanges API, so this provider has no way to "+
			"read an existing Number Range's configuration back from the tenant. Instead, write a "+
			"sapintegrationsuite_number_range resource block matching the object's current "+
			"configuration and run 'terraform apply' — SAP's documented behavior for a POST "+
			"against a Name that already exists on the tenant is unconfirmed, so verify in the "+
			"Monitor UI first whether this recreates it, conflicts, or overwrites it. See "+
			"docs/guides/runtime-stores-and-number-ranges.md.",
	)
}

type numberRangeDigitsValidator struct{}

func (v numberRangeDigitsValidator) Description(context.Context) string {
	return "must be an unsigned decimal integer (no sign, no leading zeros other than \"0\" itself)"
}

func (v numberRangeDigitsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v numberRangeDigitsValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !numberRangeDigits.MatchString(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, numberRangeInvalidValueSummary, v.Description(ctx)+numberRangeGotSuffix+req.ConfigValue.ValueString())
	}
}

type numberRangeMaxDigits15 struct{}

func (v numberRangeMaxDigits15) Description(context.Context) string {
	return "must be fewer than 15 digits (SAP's documented limit for max_value)"
}

func (v numberRangeMaxDigits15) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v numberRangeMaxDigits15) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if len(req.ConfigValue.ValueString()) >= 15 {
		resp.Diagnostics.AddAttributeError(req.Path, numberRangeInvalidValueSummary, v.Description(ctx)+numberRangeGotSuffix+req.ConfigValue.ValueString())
	}
}

type numberRangeFieldLengthValidator struct{}

func (v numberRangeFieldLengthValidator) Description(context.Context) string {
	return "must be an unsigned decimal integer between 0 and 14 (SAP's documented maximum field length)"
}

func (v numberRangeFieldLengthValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v numberRangeFieldLengthValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	s := req.ConfigValue.ValueString()
	if !numberRangeDigits.MatchString(s) {
		resp.Diagnostics.AddAttributeError(req.Path, numberRangeInvalidValueSummary, v.Description(ctx)+numberRangeGotSuffix+s)
		return
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.Cmp(big.NewInt(14)) > 0 {
		resp.Diagnostics.AddAttributeError(req.Path, numberRangeInvalidValueSummary, v.Description(ctx)+numberRangeGotSuffix+s)
	}
}
