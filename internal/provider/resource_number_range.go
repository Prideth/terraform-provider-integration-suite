package provider

import (
	"context"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
	CurrentValue          types.String `tfsdk:"current_value"`
	DeployedBy            types.String `tfsdk:"deployed_by"`
	DeployedOn            types.String `tfsdk:"deployed_on"`
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
			"SAP documents only create and update for this entity. Reading by name and deleting " +
			"were verified on a tenant in September 2026 (GET and DELETE on " +
			"NumberRanges('<name>')), so this resource detects drift, supports destroy and can be " +
			"imported by name. Create refuses to run when a number range of that name already " +
			"exists, because SAP does not document what a create on an existing name does.\n\n" +
			"The runtime counter (SAP's CurrentValue, shown as 'Next Value' in the UI) is " +
			"handled separately from the rest of this resource's configuration: see " +
			"current_value_wo below. Ordinary applies that only change description, min_value, " +
			"max_value, rotate, or field_length never send the counter, and current_value shows " +
			"its live value.",
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
				Description: "The Number Range object's name, SAP's OData key. Changing it " +
					"replaces the number range. A tenant accepted a name without special " +
					"characters (tfAccProbeNr); a request with hyphens in the name and different " +
					"values failed, so prefer plain letters and digits.",
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
					"to push to SAP, as a decimal digit string. Write-only: Terraform does not " +
					"store it; the live counter is reported in current_value instead. Sent on " +
					"Create (SAP's documented example always includes CurrentValue). On Update, " +
					"it is sent ONLY when current_value_wo_version changes; every other Update " +
					"omits CurrentValue, so that changing only description/min_value/max_value/" +
					"rotate/field_length can never reset a counter that has advanced through " +
					"EDI/EDIFACT processing.",
			},
			"current_value_wo_version": schema.StringAttribute{
				Required: true,
				Description: "An arbitrary marker (for example a counter or timestamp) that you " +
					"change to push current_value_wo to SAP on this apply. After an import, the " +
					"first apply only records the marker and does not touch the counter; change " +
					"it once more to set the counter deliberately.",
			},
			"current_value": schema.StringAttribute{
				Computed: true,
				Description: "The live counter as SAP reports it on the last read. It advances as " +
					"deployed content consumes numbers.",
			},
			"deployed_by": schema.StringAttribute{
				Computed:    true,
				Description: "User or client that last deployed the number range, as SAP reports it.",
			},
			"deployed_on": schema.StringAttribute{
				Computed:    true,
				Description: "When the number range was last deployed, RFC 3339 in UTC.",
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

	name := plan.Name.ValueString()

	// SAP does not document what a create on an existing name does. Stop
	// rather than risk overwriting a number range, and its counter, that
	// this configuration does not own.
	if _, err := r.client.GetNumberRange(ctx, name); err == nil {
		resp.Diagnostics.AddError(
			"Number range already exists",
			"A number range named "+name+" already exists on the tenant. Import it with "+
				"terraform import instead of creating it.",
		)
		return
	} else if !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to check for an existing SAP Integration Suite number range", diagnosticDetail(err))
		return
	}

	value := currentValue.ValueString()
	err := r.client.CreateNumberRange(ctx, cloudintegration.NumberRange{
		Name:         name,
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

	r.readInto(ctx, name, plan, &resp.State, &resp.Diagnostics)
}

// readInto reads the number range back and stores it, keeping the
// write-only counter marker from base.
func (r *numberRangeResource) readInto(ctx context.Context, name string, base numberRangeModel, state *tfsdk.State, diags *diag.Diagnostics) {
	nr, err := r.client.GetNumberRange(ctx, name)
	if err != nil {
		if isNotFound(err) {
			state.RemoveResource(ctx)
			return
		}
		diags.AddError("Failed to read SAP Integration Suite number range", diagnosticDetail(err))
		return
	}
	diags.Append(state.Set(ctx, numberRangeToModel(nr, base))...)
}

func numberRangeToModel(nr *cloudintegration.NumberRangeState, base numberRangeModel) numberRangeModel {
	m := base
	m.ID = types.StringValue(nr.Name)
	m.Name = types.StringValue(nr.Name)
	m.MinValue = types.StringValue(nr.MinValue)
	m.MaxValue = types.StringValue(nr.MaxValue)
	m.Description = types.StringValue(nr.Description)
	if rotate, err := strconv.ParseBool(nr.Rotate); err == nil {
		m.Rotate = types.BoolValue(rotate)
	}
	m.FieldLength = types.StringValue(nr.FieldLength)
	m.CurrentValueWO = types.StringNull()
	m.CurrentValue = stringOrNull(nr.CurrentValue)
	m.DeployedBy = stringOrNull(nr.DeployedBy)
	m.DeployedOn = odataDateToRFC3339(nr.DeployedOn)
	return m
}

func (r *numberRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state numberRangeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.readInto(ctx, state.ID.ValueString(), state, &resp.State, &resp.Diagnostics)
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

	// current_value_wo is only pushed when current_value_wo_version changed
	// since the last apply, a deliberate signal. After an import the prior
	// marker is null: that first apply only records the marker, so adopting
	// a number range in use never resets its counter.
	if !state.CurrentValueWOVersion.IsNull() && !plan.CurrentValueWOVersion.Equal(state.CurrentValueWOVersion) {
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

	r.readInto(ctx, nr.Name, plan, &resp.State, &resp.Diagnostics)
}

func (r *numberRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state numberRangeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNumberRange(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete SAP Integration Suite number range", diagnosticDetail(err))
	}
}

// ImportState imports a number range by name. current_value_wo_version stays
// null until the first apply, which then records it without touching the
// counter (see Update).
func (r *numberRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRootID(), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("name"), req.ID)...)
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
