package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apimanagementclassic"
)

// NewAPIProductResource returns a fresh resource.Resource implementation
// for sapintegrationsuite_api_product.
func NewAPIProductResource() resource.Resource {
	return &apiProductResource{}
}

type apiProductResource struct {
	client *apimanagementclassic.Client
}

type apiProductModel struct {
	ID                   types.String                        `tfsdk:"id"`
	Name                 types.String                        `tfsdk:"name"`
	Version              types.String                        `tfsdk:"version"`
	Title                types.String                        `tfsdk:"title"`
	Description          types.String                        `tfsdk:"description"`
	Scope                types.String                        `tfsdk:"scope"`
	StatusCode           types.String                        `tfsdk:"status_code"`
	IsPublished          types.Bool                          `tfsdk:"is_published"`
	IsRestricted         types.Bool                          `tfsdk:"is_restricted"`
	QuotaCount           types.Int64                         `tfsdk:"quota_count"`
	QuotaInterval        types.Int64                         `tfsdk:"quota_interval"`
	QuotaTimeUnit        types.String                        `tfsdk:"quota_time_unit"`
	APIProxyNames        []string                            `tfsdk:"api_proxy_names"`
	AdditionalProperties []apiProductAdditionalPropertyModel `tfsdk:"additional_properties"`
}

type apiProductAdditionalPropertyModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (r *apiProductResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_product"
}

func (r *apiProductResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Classic API Management API Product (APIProducts): a bundle of one " +
			"or more API Proxies published together for subscription. Backed by the API Portal's " +
			"Management.svc OData API, confirmed field-for-field against SAP's own documented Create " +
			"(POST) and Update (PUT) worked examples.\n\n" +
			"api_proxy_names is only ever sent on Create: SAP's documented Update payload never " +
			"includes the apiProxies association, so this provider treats it as immutable after " +
			"creation (RequiresReplace) rather than guess at an unconfirmed way to add or remove " +
			"proxies from an existing product. Referenced proxies are expected to already exist — " +
			"this provider does not implement sapintegrationsuite_api_proxy in this phase (see " +
			"docs/guides/classic-api-management.md), so api_proxy_names currently only accepts names " +
			"of proxies created through the SAP Integration Suite UI.\n\n" +
			"additional_properties (SAP's custom Product attributes) has its own confirmed, " +
			"independent Create/Update/Delete lifecycle (APIProductAdditionalProperties, composite " +
			"key entityId+name) and is reconciled by this resource as a set of {name, value} pairs; " +
			"SAP documents limits of 255 characters for a name, 1024 for a value, and 18 attributes " +
			"per product, none of which this provider validates itself, since SAP may change them.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Always equal to name — SAP's confirmed OData key for this entity (APIProducts('<name>')).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "The API product's name.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"version": schema.StringAttribute{
				Optional:    true,
				Description: "The product's version identifier, for example \"1\" (SAP's confirmed example value).",
			},
			"title":       schema.StringAttribute{Optional: true},
			"description": schema.StringAttribute{Optional: true},
			"scope": schema.StringAttribute{
				Optional:    true,
				Description: "SAP's confirmed example always sends this field, as an empty string when unused.",
			},
			"status_code": schema.StringAttribute{
				Optional: true,
				Description: "The product's status. The only value this provider found confirmed in " +
					"a worked example is \"PUBLISHED\"; not validated against a closed enum, since " +
					"SAP's full set of accepted values is not confirmed.",
			},
			"is_published": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the product is published (visible for subscription).",
			},
			"is_restricted": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether subscription to this product is restricted (requires approval).",
			},
			"quota_count": schema.Int64Attribute{
				Optional: true,
				Description: "Request quota count. SAP's own worked examples send -99 for \"no " +
					"quota\" in one place and null in another; this provider sends whatever is " +
					"configured (including a negative sentinel) and null when unset, without " +
					"interpreting the value itself.",
			},
			"quota_interval": schema.Int64Attribute{
				Optional:    true,
				Description: "Request quota interval, paired with quota_count and quota_time_unit.",
			},
			"quota_time_unit": schema.StringAttribute{
				Optional:    true,
				Description: "Request quota time unit, paired with quota_count and quota_interval.",
			},
			"api_proxy_names": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Names of already-existing API Proxies to associate with this product " +
					"at creation time. See the resource description above for why this is " +
					"RequiresReplace.",
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"additional_properties": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Custom name/value attributes attached to this product. Order carries no meaning, hence a set.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":  schema.StringAttribute{Required: true},
						"value": schema.StringAttribute{Required: true},
					},
				},
			},
		},
	}
}

func (r *apiProductResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func apiProductToClient(plan apiProductModel) apimanagementclassic.APIProduct {
	product := apimanagementclassic.APIProduct{
		Name:          plan.Name.ValueString(),
		Version:       plan.Version.ValueString(),
		Title:         plan.Title.ValueString(),
		Description:   plan.Description.ValueString(),
		Scope:         plan.Scope.ValueString(),
		StatusCode:    plan.StatusCode.ValueString(),
		IsPublished:   plan.IsPublished.ValueBool(),
		IsRestricted:  plan.IsRestricted.ValueBool(),
		ApiProxyNames: plan.APIProxyNames,
	}
	if !plan.QuotaCount.IsNull() {
		v := plan.QuotaCount.ValueInt64()
		product.QuotaCount = &v
	}
	if !plan.QuotaInterval.IsNull() {
		v := plan.QuotaInterval.ValueInt64()
		product.QuotaInterval = &v
	}
	if !plan.QuotaTimeUnit.IsNull() {
		v := plan.QuotaTimeUnit.ValueString()
		product.QuotaTimeUnit = &v
	}
	return product
}

func apiProductAdditionalPropertiesFromPlan(name string, props []apiProductAdditionalPropertyModel) []apimanagementclassic.APIProductAdditionalProperty {
	result := make([]apimanagementclassic.APIProductAdditionalProperty, 0, len(props))
	for _, p := range props {
		result = append(result, apimanagementclassic.APIProductAdditionalProperty{
			EntityID: name,
			Name:     p.Name.ValueString(),
			Value:    p.Value.ValueString(),
		})
	}
	return result
}

func (r *apiProductResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiProductModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateAPIProduct(ctx, apiProductToClient(plan))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Classic API Management API product", diagnosticDetail(err))
		return
	}

	for _, prop := range apiProductAdditionalPropertiesFromPlan(created.Name, plan.AdditionalProperties) {
		if err := r.client.CreateAPIProductAdditionalProperty(ctx, prop); err != nil {
			resp.Diagnostics.AddError("Failed to create Classic API Management API product additional property", diagnosticDetail(err))
			return
		}
	}

	plan.ID = types.StringValue(created.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiProductModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.client.GetAPIProduct(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Classic API Management API product", diagnosticDetail(err))
		return
	}

	state.Name = types.StringValue(found.Name)
	state.Version = stringOrNull(found.Version)
	state.Title = stringOrNull(found.Title)
	state.Description = stringOrNull(found.Description)
	state.Scope = stringOrNull(found.Scope)
	state.StatusCode = stringOrNull(found.StatusCode)
	state.IsPublished = types.BoolValue(found.IsPublished)
	state.IsRestricted = types.BoolValue(found.IsRestricted)
	if found.QuotaCount != nil {
		state.QuotaCount = types.Int64Value(*found.QuotaCount)
	} else {
		state.QuotaCount = types.Int64Null()
	}
	if found.QuotaInterval != nil {
		state.QuotaInterval = types.Int64Value(*found.QuotaInterval)
	} else {
		state.QuotaInterval = types.Int64Null()
	}
	if found.QuotaTimeUnit != nil {
		state.QuotaTimeUnit = types.StringValue(*found.QuotaTimeUnit)
	} else {
		state.QuotaTimeUnit = types.StringNull()
	}
	if len(found.ApiProxyNames) > 0 {
		state.APIProxyNames = found.ApiProxyNames
	}
	// additional_properties are not returned by GET APIProducts itself
	// (SAP's confirmed response shape does not embed them); this provider
	// leaves whatever Terraform already has in state for that attribute
	// rather than guess at a separate list-fetch call this project could
	// not confirm.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state apiProductModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateAPIProduct(ctx, apiProductToClient(plan)); err != nil {
		resp.Diagnostics.AddError("Failed to update Classic API Management API product", diagnosticDetail(err))
		return
	}

	if err := reconcileAPIProductAdditionalProperties(ctx, r.client, plan.ID.ValueString(), state.AdditionalProperties, plan.AdditionalProperties); err != nil {
		resp.Diagnostics.AddError("Failed to update Classic API Management API product additional properties", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// reconcileAPIProductAdditionalProperties diffs the prior and desired
// additional_properties sets and issues the minimal Create/Update/Delete
// calls to converge, using each confirmed operation's own composite key.
func reconcileAPIProductAdditionalProperties(ctx context.Context, client *apimanagementclassic.Client, entityID string, before, after []apiProductAdditionalPropertyModel) error {
	beforeByName := make(map[string]string, len(before))
	for _, p := range before {
		beforeByName[p.Name.ValueString()] = p.Value.ValueString()
	}
	afterByName := make(map[string]string, len(after))
	for _, p := range after {
		afterByName[p.Name.ValueString()] = p.Value.ValueString()
	}

	for name, value := range afterByName {
		oldValue, existed := beforeByName[name]
		switch {
		case !existed:
			if err := client.CreateAPIProductAdditionalProperty(ctx, apimanagementclassic.APIProductAdditionalProperty{
				EntityID: entityID, Name: name, Value: value,
			}); err != nil {
				return err
			}
		case oldValue != value:
			if err := client.UpdateAPIProductAdditionalProperty(ctx, entityID, name, value); err != nil {
				return err
			}
		}
	}
	for name := range beforeByName {
		if _, stillPresent := afterByName[name]; !stillPresent {
			if err := client.DeleteAPIProductAdditionalProperty(ctx, entityID, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *apiProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiProductModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAPIProduct(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete Classic API Management API product", diagnosticDetail(err))
	}
}

func (r *apiProductResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}
