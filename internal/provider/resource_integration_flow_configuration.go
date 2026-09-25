package provider

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// NewIntegrationFlowConfigurationResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_integration_flow_configuration.
func NewIntegrationFlowConfigurationResource() resource.Resource {
	return &integrationFlowConfigurationResource{}
}

type integrationFlowConfigurationResource struct {
	client *cloudintegration.Client
}

type integrationFlowConfigurationModel struct {
	ID          types.String `tfsdk:"id"`
	FlowID      types.String `tfsdk:"flow_id"`
	FlowVersion types.String `tfsdk:"flow_version"`
	Parameters  types.Map    `tfsdk:"parameters"`
}

func (r *integrationFlowConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_flow_configuration"
}

func (r *integrationFlowConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets externalized parameters of an integration flow, the values a flow " +
			"exposes for configuration without editing its model (receiver hosts, endpoint " +
			"addresses, credential names and so on). Only the keys listed in parameters are " +
			"managed; every other parameter keeps its value. Backed by the Configurations of " +
			"IntegrationDesigntimeArtifacts in the Integration Content OData V2 API. Changed values " +
			"take effect at runtime only after the flow is redeployed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same as flow_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"flow_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the integration flow.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"flow_version": schema.StringAttribute{
				Required: true,
				Description: "Design-time version whose parameters are set, typically " +
					"sapintegrationsuite_integration_flow.<name>.version. When it changes, all " +
					"managed parameters are written again into the new version.",
			},
			"parameters": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Parameter keys and the values to set, as strings. Each key must exist " +
					"as an externalized parameter of the flow version. The parameter's existing data " +
					"type is kept.",
			},
		},
	}
}

func (r *integrationFlowConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *integrationFlowConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan integrationFlowConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = plan.FlowID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *integrationFlowConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state integrationFlowConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configs, err := r.client.ListIntegrationFlowConfigurations(ctx, state.FlowID.ValueString(), state.FlowVersion.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read integration flow configuration", diagnosticDetail(err))
		return
	}

	var managed map[string]string
	resp.Diagnostics.Append(state.Parameters.ElementsAs(ctx, &managed, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current := configurationValues(configs)
	observed := make(map[string]string, len(managed))
	for key := range managed {
		if value, ok := current[key]; ok {
			observed[key] = value
		}
	}

	params, diags := types.MapValueFrom(ctx, types.StringType, observed)
	resp.Diagnostics.Append(diags...)
	state.Parameters = params
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *integrationFlowConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan integrationFlowConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = plan.FlowID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete only forgets the parameters. SAP has no operation to remove or reset
// an externalized parameter, so the values stay as they are.
func (r *integrationFlowConfigurationResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Integration flow parameters left unchanged",
		"SAP offers no API to reset externalized parameters. Terraform stopped managing them, "+
			"and they keep their current values on the flow.")
}

// ImportState accepts "<flow_id>/<flow_version>" and adopts every parameter the
// version currently has. Parameters later removed from the configuration are
// simply no longer managed.
func (r *integrationFlowConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	flowID, version, err := splitCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected \"<flow_id>/<flow_version>\": "+err.Error())
		return
	}

	configs, err := r.client.ListIntegrationFlowConfigurations(ctx, flowID, version)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read integration flow configuration", diagnosticDetail(err))
		return
	}
	params, diags := types.MapValueFrom(ctx, types.StringType, configurationValues(configs))
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRootID(), flowID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("flow_id"), flowID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("flow_version"), version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parameters"), params)...)
}

// apply writes every planned parameter, keeping each parameter's DataType.
// Unknown keys are reported together before anything is written.
func (r *integrationFlowConfigurationResource) apply(ctx context.Context, plan integrationFlowConfigurationModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var desired map[string]string
	diags.Append(plan.Parameters.ElementsAs(ctx, &desired, false)...)
	if diags.HasError() {
		return diags
	}

	flowID, version := plan.FlowID.ValueString(), plan.FlowVersion.ValueString()
	configs, err := r.client.ListIntegrationFlowConfigurations(ctx, flowID, version)
	if err != nil {
		diags.AddError("Failed to read integration flow configuration", diagnosticDetail(err))
		return diags
	}
	dataTypes := make(map[string]string, len(configs))
	for _, c := range configs {
		dataTypes[c.ParameterKey] = c.DataType
	}

	keys := make([]string, 0, len(desired))
	var unknown []string
	for key := range desired {
		keys = append(keys, key)
		if _, ok := dataTypes[key]; !ok {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(keys)
	if len(unknown) > 0 {
		sort.Strings(unknown)
		available := make([]string, 0, len(configs))
		for _, c := range configs {
			available = append(available, c.ParameterKey)
		}
		diags.AddAttributeError(path.Root("parameters"), "Unknown externalized parameter",
			fmt.Sprintf("Integration flow %q version %q has no externalized parameter %s. Available: %s.",
				flowID, version, strings.Join(unknown, ", "), strings.Join(available, ", ")))
		return diags
	}

	for _, key := range keys {
		if err := r.client.UpdateIntegrationFlowConfiguration(ctx, flowID, version, key, desired[key], dataTypes[key]); err != nil {
			diags.AddError(fmt.Sprintf("Failed to set parameter %q of integration flow %q", key, flowID), diagnosticDetail(err))
			return diags
		}
	}
	return diags
}

func configurationValues(configs []cloudintegration.IntegrationFlowConfiguration) map[string]string {
	values := make(map[string]string, len(configs))
	for _, c := range configs {
		values[c.ParameterKey] = c.ParameterValue
	}
	return values
}
